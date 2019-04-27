-- luacheck: globals lmdb
local utils = lmdb.utils
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
local count = lmdb.count("*")
assert(count == 1)
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


local function test_utils_to_hex()
    local ok, err = pcall(utils.to_hex)
    assert(not ok and err == "at least one argument expected")

    local data = {
        {"f", "MY======"},
        {"fo", "MZXQ===="},
        {"foo", "MZXW6==="},
        {"Twas brillig, and the slithy toves",
            "KR3WC4ZAMJZGS3DMNFTSYIDBNZSCA5DIMUQHG3DJORUHSIDUN53GK4Y="},
    }
    for _, t in ipairs(data) do
        local input = t[1]
        local exp = t[2]
        local act = utils.to_hex(input)
        assert(exp == act,
               "input " .. input .. " expected " .. exp .. " got " .. act)
    end
end

test_utils_to_hex()
