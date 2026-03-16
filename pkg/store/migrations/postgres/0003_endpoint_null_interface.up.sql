ALTER TABLE "application_endpoints" DROP CONSTRAINT IF EXISTS "port_protocol_addr_network_interface_id";
CREATE UNIQUE INDEX IF NOT EXISTS "port_protocol_addr_network_interface_id" ON "application_endpoints" ("port", "protocol", "addr", COALESCE("network_interface_id", 0));
