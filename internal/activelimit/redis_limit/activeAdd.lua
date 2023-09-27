-- 限流对象
local key = KEYS[1]
-- 最大限流个数
local maxActive = tonumber(ARGV[1])

local countActive = tonumber(redis.call('INCR', key))
if countActive <= maxActive then
    return "false"
else
    redis.call('DECR', key)
    return "true"
end


