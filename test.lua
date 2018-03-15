-- luacheck: globals lmdb
local keys = lmdb.keys("*")
assert(next(keys) == nil)

-- test error handling
assert(lmdb.non_exist == nil)
local res, err = lmdb.set("key")
assert(res == nil)
assert(err == "wrong number of arguments for 'set' command", err)
local key = nil
local ok  = pcall(lmdb.set, key, 1)
assert(not ok)
local ok, err = lmdb.get("non_exist")
assert(not ok)
assert(err == "not found", err)

assert(lmdb.set("key", 1))
assert(lmdb.exists("key"))
assert(lmdb.get("key") == "1")
assert(lmdb.put("empty", ""))
assert(lmdb.get("empty") == "")

local keys = lmdb.keys("*")
assert(#keys == 2)
assert(1, lmdb.del("empty"))
local keys = lmdb.keys("*")
assert(#keys == 1)
assert(keys[1] == "key")

assert(1, lmdb.del("key"))
assert(not lmdb.exists("key"))

assert(lmdb.put("db", "key", 1))
assert(lmdb.exists("db", "key"))

local ok, err = lmdb.set("db", 2)
assert(not ok)
assert(err == "incompatible operation", err)

local stat = lmdb.stat()
for k, v in pairs(stat) do
    print(k, tostring(v))
end
assert(stat.psize > 0)
assert(stat.entries == 1)
assert(stat.depth == 1)

lmdb.set("a", 1)
lmdb.set("b", 2)
-- dumping all values
local keys = lmdb.keys('*')
for _, k in ipairs(keys) do
    print(lmdb.get(k))
end

-- delete all matched keys
for _, k in ipairs(keys) do
    lmdb.del(k)
end
assert(next(lmdb.keys('*')) == nil)
