-- Github Webhook Event
-- Ref: https://docs.github.com/en/developers/webhooks-and-events/webhooks

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
local event = req:header("X-GitHub-Event")
local title = data["repository"]["full_name"]

if event == "push" then
	msg = string.format("%s - new commit:\n%s", title, data["head_commit"]["message"])
else
	msg = string.format("%s\n%s %s", title, event, data["action"])
end

local ret = ctx:send({
	title = "Github",
	text = msg,
	sound = req:query("sound"),
})

return 200, ret
