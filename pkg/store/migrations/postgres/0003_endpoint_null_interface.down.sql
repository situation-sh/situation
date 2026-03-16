DROP INDEX IF EXISTS "port_protocol_addr_network_interface_id";
ALTER TABLE "application_endpoints" ADD CONSTRAINT "port_protocol_addr_network_interface_id" UNIQUE ("port", "protocol", "addr", "network_interface_id");
