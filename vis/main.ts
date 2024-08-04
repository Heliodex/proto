// quick scrappy js app to visualise connections

import type { ServerWebSocket } from "bun"

const sockets = new Map<string, ServerWebSocket<unknown>>() // probably dont work w/ more than 1

Bun.serve({
	async fetch(req, server) {
		const url = new URL(req.url)

		if (url.pathname === "/realtime") {
			const upgraded = server.upgrade(req)
			if (!upgraded)
				return new Response("Upgrade failed", { status: 400 })
		} else if (url.pathname === "/notify") {
			const body = await req.text()
			for (const [, v] of sockets) v.send(body)
			return new Response("yeah")
		} else if (url.pathname == "/script.js")
			return new Response(Bun.file("./script.js"))
		else if (req.method === "GET")
			return new Response(Bun.file("./index.html"))
		return new Response("Not Found", { status: 404 })
	},

	websocket: {
		open: ws => {
			if (sockets.has(ws.remoteAddress)) ws.close()
			sockets.set(ws.remoteAddress, ws)
			console.log("Client connected", sockets.size)
		},
		message: () => {},
		close: ws => {
			sockets.delete(ws.remoteAddress)
			console.log("Client disconnected", sockets.size)
		},
	},
})

console.log("Server up")
