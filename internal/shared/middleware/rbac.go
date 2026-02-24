package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/platform/cache"
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/modules/users"
)

// RBACConfig agrupa los repositorios necesarios para la verificación de permisos
type RBACConfig struct {
	UserRepo       users.UserRepository
	RoleRepo       roles.RoleRepository
	PermissionRepo permissions.PermissionRepository
	ResourceRepo   resources.ResourceRepository
	Cache          cache.Cache // Optional cache for RBAC data
}

// RBACMiddleware verifica que el usuario autenticado tenga permiso para
// ejecutar la acción HTTP sobre el recurso solicitado.
//
// Requiere que JWTMiddleware haya sido ejecutado antes (user_id en contexto).
//
// Flujo (con cache):
//  1. Check cache for user permissions → hit: skip to step 4
//  2. Obtener user → role_ids
//  3. Obtener roles WHERE _id IN role_ids → aplanar permissions_ids
//  4. Obtener permissions WHERE _id IN permissions_ids AND action = método HTTP
//  5. Verificar si algún resource de esos permissions tiene name = segmento de ruta
//  6. Cache result for future requests
func RBACMiddleware(cfg RBACConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userID := sharedAuth.GetUserID(c)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
			})
			return
		}

		// Determinar acción desde el método HTTP (get, post, put, patch, delete)
		action := permissions.Action(strings.ToLower(c.Request.Method))
		if !permissions.ValidActions[action] {
			c.Next()
			return
		}

		// Extraer nombre del recurso desde la ruta registrada en Gin
		// Ej: /api/appointments/:id → "appointments"
		resourceName := extractResourceName(c.FullPath())
		if resourceName == "" {
			c.Next()
			return
		}

		// Try cache first if enabled
		if cfg.Cache != nil && cfg.Cache.IsEnabled() {
			allowed, err := checkCachePermission(ctx, cfg.Cache, userID, resourceName, string(action))
			if err == nil {
				if allowed {
					c.Next()
					return
				}
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "access denied",
				})
				return
			}
		}

		// 1. Obtener usuario y sus role_ids
		user, err := cfg.UserRepo.FindByID(ctx, userID)
		if err != nil || len(user.RoleIds) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied",
			})
			return
		}

		// 2. Obtener roles y aplanar todos sus permissions_ids
		userRoles, err := cfg.RoleRepo.FindByIDs(ctx, user.RoleIds)
		if err != nil || len(userRoles) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied",
			})
			return
		}

		allPermissionIDs := flattenPermissionIDs(userRoles)
		if len(allPermissionIDs) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied",
			})
			return
		}

		// 3. Filtrar permisos que coincidan con la acción solicitada
		matchingPerms, err := cfg.PermissionRepo.FindByIDsAndAction(ctx, allPermissionIDs, action)
		if err != nil || len(matchingPerms) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied",
			})
			return
		}

		// Extraer los resource_ids de los permisos coincidentes
		resourceIDs := make([]primitive.ObjectID, 0, len(matchingPerms))
		for _, p := range matchingPerms {
			resourceIDs = append(resourceIDs, p.ResourceId)
		}

		// 4. Verificar si algún resource con ese nombre está en los resource_ids permitidos
		allowed, err := cfg.ResourceRepo.FindByIDsAndName(ctx, resourceIDs, resourceName)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied",
			})
			return
		}

		// Cache the permission result
		if cfg.Cache != nil && cfg.Cache.IsEnabled() {
			_ = cachePermissionResult(ctx, cfg.Cache, userID, resourceName, string(action), allowed)
		}

		c.Next()
	}
}

// extractResourceName obtiene el nombre del recurso desde la ruta registrada en Gin.
// Devuelve el último segmento que no sea un parámetro de ruta (no empieza con ':').
// Ejemplos:
//   - /api/appointments       → "appointments"
//   - /api/appointments/:id   → "appointments"
//   - /api/tenants/:id/baths  → "baths"
func extractResourceName(fullPath string) string {
	segments := strings.Split(fullPath, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		seg := segments[i]
		if seg != "" && !strings.HasPrefix(seg, ":") {
			return seg
		}
	}
	return ""
}

// flattenPermissionIDs aplana y deduplica los permissions_ids de todos los roles del usuario
func flattenPermissionIDs(userRoles []*roles.Role) []primitive.ObjectID {
	seen := make(map[primitive.ObjectID]struct{})
	result := make([]primitive.ObjectID, 0)
	for _, role := range userRoles {
		for _, pid := range role.PermissionsIds {
			if _, ok := seen[pid]; !ok {
				seen[pid] = struct{}{}
				result = append(result, pid)
			}
		}
	}
	return result
}

// checkCachePermission checks if a user has permission for a resource/action in cache
// Returns (allowed, nil) on cache hit, (false, ErrCacheMiss) on cache miss
func checkCachePermission(ctx context.Context, c cache.Cache, userID, resourceName, action string) (bool, error) {
	key := fmt.Sprintf(cache.CacheKeyUserPerms, userID, fmt.Sprintf("%s:%s", resourceName, action))
	val, err := c.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

// cachePermissionResult caches the permission check result
// allowed=true is stored as "1", allowed=false as "0"
func cachePermissionResult(ctx context.Context, c cache.Cache, userID, resourceName, action string, allowed bool) error {
	key := fmt.Sprintf(cache.CacheKeyUserPerms, userID, fmt.Sprintf("%s:%s", resourceName, action))
	value := "0"
	if allowed {
		value = "1"
	}
	return c.Set(ctx, key, value, cache.CacheDefaultTTL)
}
