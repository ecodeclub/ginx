-- 限流对象
local key = KEYS[1]

local curCnt = tonumber(redis.call('DECR', key))
if curCnt < 0 then
    redis.call('Set', key, 0)
    return "false"
else
    return "true"
end



