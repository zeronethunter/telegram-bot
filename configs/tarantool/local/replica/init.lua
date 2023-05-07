-- instance file for the replica
box.cfg {
    listen = 3302,
    replication = { 'replicator:password@localhost:3301', -- master URI
                    'replicator:password@localhost:3302' }, -- replica URI
    read_only = true,
}
box.once("schema", function()
    box.schema.user.create('replicator', { password = 'password' })
    box.schema.user.grant('replicator', 'read,write,execute', 'universe', nil)
    box.schema.space.create("credentials")
    box.space.credentials:create_index("primary", { type = "tree", parts = { 1, "unsigned", 2, "string" } })
    box.space.credentials:format({
        { name = 'user_id', type = 'unsigned' },
        { name = 'service_name', type = 'string' },
        { name = 'login', type = 'string', is_nullable = true },
        { name = 'password', type = 'string', is_nullable = true }
    })

    box.schema.space.create("state")
    box.space.state:create_index("primary", { type = "tree", parts = { 1, "unsigned" } })
    box.space.state:format({
        { name = 'user_id', type = 'unsigned' },
        { name = 'state', type = 'string' },
        { name = 'last_service', type = 'string', is_nullable = true }
    })
    print('box.once executed on replica')
end)
