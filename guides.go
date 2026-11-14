package main

// Guide is a longer page about how the tools fit together. Its text is the
// template templates/guide-<Slug>.html.
type Guide struct {
	Slug    string
	Title   string
	Summary string
}

var guides = []Guide{
	{
		Slug:    "tailscale",
		Title:   "Tailscale for the vt100 stack",
		Summary: "A personal tailnet, the policy file, servers that join on first boot, and Funnel for testing on a laptop.",
	},
}
