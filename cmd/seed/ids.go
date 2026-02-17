package main

import "go.mongodb.org/mongo-driver/bson/primitive"

func mustOID(hex string) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(hex)
	return id
}

var (
	// Tenant IDs
	tenant1ID = mustOID("65f0000000000000001111a1") // Clínica Veterinaria Vida Animal - Medellín
	tenant2ID = mustOID("65f0000000000000001111a2") // Centro Veterinario Patitas Felices - Bogotá

	// Owner IDs (administradores de cada clínica)
	owner1ID = mustOID("65f0000000000000002111a1") // Dr. Santiago Herrera - Vida Animal
	owner2ID = mustOID("65f0000000000000002111a2") // Dra. Valentina Morales - Patitas Felices

	// Staff Tenant 1 - Vida Animal
	staffT1VetID          = mustOID("65f0000000000000002111b1") // Dra. Laura Quintero - Veterinaria
	staffT1ReceptionistID = mustOID("65f0000000000000002111b2") // Camila Torres - Recepcionista
	staffT1AssistantID    = mustOID("65f0000000000000002111b3") // Miguel Ángel López - Auxiliar
	staffT1AccountantID   = mustOID("65f0000000000000002111b4") // Roberto Castillo - Contador

	// Staff Tenant 2 - Patitas Felices
	staffT2VetID          = mustOID("65f0000000000000002111c1") // Dr. Andrés Ruiz - Veterinario
	staffT2ReceptionistID = mustOID("65f0000000000000002111c2") // Juliana Ramírez - Recepcionista
	staffT2AssistantID    = mustOID("65f0000000000000002111c3") // Felipe García - Auxiliar
)
