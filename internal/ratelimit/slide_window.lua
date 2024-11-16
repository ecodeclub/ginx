-- 限流对象
local key = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 阈值
local threshold = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
-- 唯一ID, 用于解决同一时间内多个请求只统计一次的问题
-- SEE: issue #27
local uid = ARGV[4]
-- 窗口的起始时间
local min = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
-- local cnt = redis.call('ZCOUNT', key, min, '+inf')
if cnt >= threshold then
    -- 执行限流
    return "true"
else
    -- score 设置为当前时间, member 设置为唯一id
    redis.call('ZADD', key, now, uid)
    redis.call('PEXPIRE', key, window)
    return "false"
end