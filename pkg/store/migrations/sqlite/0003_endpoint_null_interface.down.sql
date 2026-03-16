DROP INDEX IF EXISTS "port_protocol_addr_network_interface_id";
CREATE UNIQUE INDEX IF NOT EXISTS "port_protocol_addr_network_interface_id" ON "application_endpoints" ("port", "protocol", "addr", "network_interface_id");
