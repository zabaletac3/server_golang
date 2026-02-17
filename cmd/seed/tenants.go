package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/tenant"
)

type tenantSeedData struct {
	id             primitive.ObjectID
	ownerID        primitive.ObjectID
	name           string
	commercialName string
	nit            string
	email          string
	phone          string
	address        string
	domain         string
	planName       string
	usersLimit     int
	storageLimitMB int
}

var tenantSeeds = []tenantSeedData{
	{
		id:             tenant1ID,
		ownerID:        owner1ID,
		name:           "Clínica Veterinaria Vida Animal",
		commercialName: "Vida Animal",
		nit:            "900123456-7",
		email:          "admin@vidaanimal.com.co",
		phone:          "+57 604 345 6789",
		address:        "Calle 10 #43-21, El Poblado, Medellín, Antioquia",
		domain:         "vidaanimal",
		planName:       "Pro",
		usersLimit:     5,
		storageLimitMB: 10240, // 10 GB
	},
	{
		id:             tenant2ID,
		ownerID:        owner2ID,
		name:           "Centro Veterinario Patitas Felices",
		commercialName: "Patitas Felices",
		nit:            "901234567-8",
		email:          "admin@patitasfelices.com.co",
		phone:          "+57 601 234 5678",
		address:        "Carrera 15 #93-47, Chapinero, Bogotá, D.C.",
		domain:         "patitasfelices",
		planName:       "Empresarial",
		usersLimit:     20,
		storageLimitMB: 51200, // 50 GB
	},
}

func (s *SeedService) SeedTenants(ctx context.Context, planMap map[string]primitive.ObjectID) error {
	existing, err := s.tenantRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		s.logger.Info("tenants already exist, skipping seed")
		return nil
	}

	trialEnd := time.Now().AddDate(0, 0, 30)
	now := time.Now()

	for _, ts := range tenantSeeds {
		planID := planMap[ts.planName] // zero ObjectID si no se encuentra

		t := &tenant.Tenant{
			ID:                   ts.id,
			OwnerID:              ts.ownerID,
			Name:                 ts.name,
			CommercialName:       ts.commercialName,
			IdentificationNumber: ts.nit,
			Industry:             "veterinary",
			Email:                ts.email,
			Phone:                ts.phone,
			Address:              ts.address,
			Country:              "Colombia",
			Domain:               ts.domain,
			TimeZone:             "America/Bogota",
			Currency:             "COP",
			Subscription: tenant.TenantSubscription{
				PlanID:        planID,
				BillingStatus: "trial",
				TrialEndsAt:   &trialEnd,
				MRR:           0,
			},
			Usage: tenant.TenantUsage{
				UsersCount:     0,
				UsersLimit:     ts.usersLimit,
				StorageUsedMB:  0,
				StorageLimitMB: ts.storageLimitMB,
				LastResetDate:  now,
			},
			Status:    tenant.Trial,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.tenantRepo.Create(ctx, t); err != nil {
			s.logger.Error("failed to create tenant", "name", ts.name, "error", err)
			return err
		}
		s.logger.Info("tenant created", "name", ts.name, "plan", ts.planName, "domain", ts.domain)
	}

	return nil
}
