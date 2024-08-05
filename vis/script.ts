const circle = document.getElementById("circle")
if (!circle) throw new Error("no circle")

const nodes: HTMLElement[] = [
	...(document.getElementsByClassName("node") as unknown as any[]),
]

const maxNodes = 5
// there's already 1
for (let i = 0; i < maxNodes - 1; i++) {
	const currentNode = nodes[i].cloneNode(true) as HTMLElement
	nodes.push(currentNode)
	circle.appendChild(currentNode)
}

const getRealNode = (i: number) =>
	nodes[i].getElementsByClassName("nodenode")[0] as HTMLElement

let i = 0

const svg = document.getElementById("svg")
if (!svg) throw new Error("no svg")
const path = document.getElementById("path") as unknown as SVGPathElement
if (!path) throw new Error("no path")

const paths = new Map<
	string,
	{
		n1: HTMLElement
		n2: HTMLElement
		path: SVGPathElement
	}
>()

const getPath = (p1: DOMRect, p2: DOMRect) => `M${p1.x} ${p1.y} ${p2.x} ${p2.y}`

const getCentre = (d: DOMRect) =>
	new DOMRect(d.x + d.width / 2, d.y + d.height / 2)

function update() {
	requestAnimationFrame(update)
	const diff = 360 / nodes.length
	// type shenanigans
	for (let j = 0; j < nodes.length; j++)
		nodes[j].style.rotate = `${(i + diff * j) % 360}deg`

	for (const [, p] of paths) {
		let p1 = new DOMRect(0, 0)
		let p2 = new DOMRect(0, 0)
		if (p.n1) p1 = getCentre(p.n1.getBoundingClientRect())
		if (p.n2) p2 = getCentre(p.n2.getBoundingClientRect())

		p.path.setAttribute("d", getPath(p1, p2))
	}
	i += 0.1
}
requestAnimationFrame(update)

// thanks jrchibald archibald
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

	const realnode = getRealNode(data.Address - 1)

	switch (data.Type) {
		case "Send":
			const rand = Math.random().toString().slice(2)
			const newPath = path.cloneNode(true) as SVGPathElement
			newPath.id = rand
			paths.set(rand, {
				path: newPath,
				n1: realnode,
				n2: getRealNode(data.To - 1),
			})
			svg.appendChild(newPath)

			realnode.style.transition = ""
			realnode.style.background = "magenta"

			newPath.style.transition = ""
			newPath.style.stroke = "#666"
			doubleRaf(() => {
				realnode.style.transition = "background 0.5s"
				realnode.style.background = "#666"

				newPath.style.transition = "stroke 0.5s"
				newPath.style.stroke = "transparent"
				setTimeout(() => {
					svg.removeChild(newPath)
					paths.delete(rand) // memory leak plaster
				}, 500) // watch this remove too early because sync
			})
			break
		case "Receive":
			realnode.style.transition = ""
			realnode.style.background = "lime"
			doubleRaf(() => {
				realnode.style.transition = "background 0.5s"
				realnode.style.background = "#666"
			})
			break
	}
}
socket.onclose = () => {
	console.log("closed")
}
