CREATE INDEX IF NOT EXISTS "network_interface_ip" ON "network_interfaces" ("ip");
CREATE INDEX IF NOT EXISTS "network_interface_mac" ON "network_interfaces" ("mac");
CREATE INDEX IF NOT EXISTS "application_endpoint_protocol_port" ON "application_endpoints" ("protocol", "port");