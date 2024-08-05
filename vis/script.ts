const circle = document.getElementById("circle")
if (!circle) throw new Error("no circle")

const nodes: HTMLElement[] = [
	...(document.getElementsByClassName("node") as unknown as any[]),
]

const maxNodes = 2
// there's already 1
for (let i = 0; i < maxNodes - 1; i++) {
	const currentNode = nodes[i].cloneNode(true) as HTMLElement
	nodes.push(currentNode)
	circle.appendChild(currentNode)
}

const getRealNode = (i: number) =>
	nodes[i].getElementsByClassName("nodenode")[0] as HTMLElement

let i = 0

function update() {
	requestAnimationFrame(update)
	const diff = 360 / nodes.length
	// type shenanigans
	for (let j = 0; j < nodes.length; j++)
		nodes[j].style.rotate = `${(i + diff * j) % 360}deg`
	i++
}
requestAnimationFrame(update)

const doubleRaf = (f: () => void) =>
	requestAnimationFrame(() => {
		requestAnimationFrame(f)
	})

// socketzz

const socket = new WebSocket("/realtime")

socket.onopen = () => {
	console.log("opened")
	socket.send("hi")

	setInterval(() => {
		socket.send("keepalive")
	}, 30e3)
}
socket.onmessage = msg => {
	const data = JSON.parse(msg.data)
	console.log(data)

	let colour = "#fff"
	if (data.Type === "Send") colour = "magenta"
	else if (data.Type === "Receive") colour = "lime"

	const realnode = getRealNode(data.Address - 1)
	realnode.style.transition = ""
	realnode.style.background = colour
	doubleRaf(() => {
		realnode.style.transition = "background 0.5s"
		realnode.style.background = "#666"
	})
}
socket.onclose = () => {
	console.log("closed")
}
