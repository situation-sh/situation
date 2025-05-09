// walk through all div.codehilite
// and add a copy button to each of them
for (const ch of document.querySelectorAll("div.codehilite")) {
	// the copy button can be found anywhere in the ui.shadcn website
	// ex: https://ui.shadcn.com/docs/theming
	const button = document.createElement("button");
	button.setAttribute(
		"class",
		"cursor-pointer inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 absolute right-2.5 top-2 z-10 h-6 w-6 text-zinc-50 hover:bg-zinc-700 hover:text-zinc-50 [&_svg]:h-3 [&_svg]:w-3",
	);

	const span = document.createElement("span");
	span.setAttribute("class", "sr-only");
	span.innerText = "Copy";
	button.appendChild(span);

	const icon = clipboardIcon();
	button.appendChild(icon);
	button.onclick = onCodeCopy;
	ch.appendChild(button);
}
