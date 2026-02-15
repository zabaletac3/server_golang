package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
}

// RBACMiddleware verifica que el usuario autenticado tenga permiso para
// ejecutar la acción HTTP sobre el recurso solicitado.
//
// Requiere que JWTMiddleware haya sido ejecutado antes (user_id en contexto).
//
// Flujo (4 queries a MongoDB):
//  1. Obtener user → role_ids
//  2. Obtener roles WHERE _id IN role_ids → aplanar permissions_ids
//  3. Obtener permissions WHERE _id IN permissions_ids AND action = método HTTP
//  4. Verificar si algún resource de esos permissions tiene name = segmento de ruta
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
