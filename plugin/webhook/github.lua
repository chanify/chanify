-- Github Webhook Event

local hex = require 'hex'
local json = require 'json'
local crypto = require'crypto'

local seckey = ctx:env("SECRET_TOKEN")
local req = ctx:request()
local sign = hex.decode(req:header("X-HUB-SIGNATURE-256"):sub(8))

if not crypto.equal(crypto.hmac("sha256", seckey, req:body()), sign) then
	return 401
end

local data = json.decode(req:body())

local ret = ctx:send("Github " .. data["repository"]["full_name"])

return 200, ret
