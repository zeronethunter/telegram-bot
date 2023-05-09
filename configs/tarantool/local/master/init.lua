-- instance file for the master
box.cfg {
    listen = 3301,
    replication = { 'replicator:' .. os.getenv("TARANTOOL_PASSWORD") .. '@localhost:3301', -- master URI
                    'replicator:' .. os.getenv("TARANTOOL_PASSWORD") .. '@localhost:3302' }, -- replica URI
    read_only = false,
}
box.once("schema", function()
    box.schema.user.create('replicator', { password = os.getenv("TARANTOOL_PASSWORD") })
    box.schema.user.grant('replicator', 'read,write,execute', 'universe', nil)
    box.schema.space.create("users")
    box.space.users:create_index("primary", { type = "tree", parts = { 1, "unsigned", 2, "string" } })
    box.space.users:format({
        { name = 'user_id', type = 'unsigned' },
        { name = 'token', type = 'string' },
    })

    box.schema.space.create("credentials")
    box.space.credentials:create_index("primary", { type = "tree", parts = { 1, "unsigned", 2, "string" } })
    box.space.credentials:format({
        { name = 'user_id', type = 'unsigned' },
        { name = 'service_name', type = 'string' },
        { name = 'login', type = 'string', is_nullable = true },
        { name = 'password', type = 'string', is_nullable = true },
    })

    box.schema.space.create("state")
    box.space.state:create_index("primary", { type = "tree", parts = { 1, "unsigned" } })
    box.space.state:format({
        { name = 'user_id', type = 'unsigned' },
        { name = 'state', type = 'string' },
        { name = 'last_service', type = 'string', is_nullable = true }
    })
    print('box.once executed on master')
end)
