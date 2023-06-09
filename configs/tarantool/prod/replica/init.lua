-- instance file for the replica
box.cfg {
    listen = 3301,
    replication = { 'replicator:' .. os.getenv("TARANTOOL_PASSWORD") .. '@tarantool-master:3301', -- master URI
                    'replicator:' .. os.getenv("TARANTOOL_PASSWORD") .. '@tarantool-replica:3301' }, -- replica URI
    read_only = true,
}
box.once("schema", function()
    local secret = os.getenv("TARANTOOL_PASSWORD")

    box.schema.user.create('replicator', { password = secret })
    box.schema.user.grant('replicator', 'read,write,execute', 'universe', nil)
    box.schema.space.create("users")
    box.space.users:create_index("primary", { type = "tree", parts = { 1, "unsigned" } })
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
    print('box.once executed on replica')
end)
