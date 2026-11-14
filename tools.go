package main

import "html/template"

// githubOwner is the GitHub organization the tools live in.
const githubOwner = "vt100io"

// Tool is one tool on the site. Fields of type template.HTML are written
// here, by me, and may contain markup such as <code>.
type Tool struct {
	Name    string
	Tagline template.HTML
	LeansOn []string
	Install string // the command that gets it
	Demo    string // path under /static, optional
	DemoAlt string
	Why     []template.HTML
	Not     []Not
	Steps   []template.HTML
}

// Not is something a tool deliberately isn't.
type Not struct{ Title, Text string }

func (t Tool) Repo() string { return "https://github.com/" + githubOwner + "/" + t.Name }

var tools = []Tool{
	{
		Name:    "punchcard",
		Tagline: `A one-binary CI server that runs the same <code>mise run ci</code> you run on your laptop, on a machine that remembers.`,
		LeansOn: []string{"mise", "GitHub App"},
		Install: "go install github.com/" + githubOwner + "/punchcard@latest",
		Demo:    "/static/img/punchcard-tui.gif",
		DemoAlt: "punchcard tui: a build arrives and streams its log, a newer push replaces a queued one, and a failed build is opened",
		Why: []template.HTML{
			`Hosted CI runners start from nothing on every build: a fresh machine, fresh tool installs, cold caches. Punchcard keeps a workspace per repo on a machine I own. When I push, it fetches, resets to the exact commit and runs the repo's <code>ci</code> task.`,
			`There is no CI language. The pipeline is a mise task in the repo, so if it passes on my laptop it passes in CI, and a CI failure can be debugged on my own machine.`,
			`A GitHub App posts the checks. A repo opts in by having a <code>ci</code> task; there is no list of repos to keep.`,
		},
		Not: []Not{
			{"A sandbox.", "Anyone who can open a PR runs code on the machine."},
			{"Parallel.", "One build at a time keeps workspaces safe."},
			{"A platform.", "No agents, runners, plugins or marketplace."},
		},
		Steps: []template.HTML{
			`Create a GitHub App: contents read, commit statuses write, push and pull request events.`,
			`Deploy it to a server with <code>unit deploy</code> and a <code>domain</code> in its <code>unit.toml</code>. Caddy gets the certificate, and <code>https://&lt;domain&gt;/</code> is the App's webhook URL. To try it on a laptop first, run <code>punchcard</code> and publish it with <code>tailscale funnel --bg 8081</code>.`,
			`Add <code>[tasks.ci]</code> to a repo's mise.toml and push.`,
			`Watch it: <code>punchcard tui</code> on the machine it runs on.`,
		},
	},
	{
		Name:    "unit",
		Tagline: `Deploys statically compiled binaries as hardened systemd units over SSH. No containers, no daemons, just PID 1.`,
		LeansOn: []string{"systemd", "SSH"},
		Install: "go install github.com/" + githubOwner + "/unit/cmd/unit@latest",
		Why: []template.HTML{
			`A statically compiled binary runs on any Linux host as-is. A container adds a daemon to manage, a registry to push to and a base image to pull, and buys nothing for a binary that has no runtime dependencies.`,
			`systemd already runs as PID 1 on every Linux server, with the isolation, secrets, resource limits, socket activation and timers you'd want. The gap is ergonomics: a hardened unit is 40+ lines before anything is specific to your app. unit takes a short description of what the app needs and writes the units correctly every time.`,
		},
		Not: []Not{
			{"A container runtime.", "It deploys static binaries, nothing else."},
			{"A daemon.", "It runs on your machine, does its work over SSH and exits."},
		},
	},
	{
		Name:    "smed",
		Tagline: `Configures machines from plain Go: it reads each machine, shows a plan, then applies it. No interpreter on the target, no state file anywhere.`,
		LeansOn: []string{"Tailscale", "SSH"},
		Install: "go get github.com/" + githubOwner + "/smed",
		Why: []template.HTML{
			`Ansible looks at a machine before it acts, but shows no plan. Terraform shows a plan, but against a state file instead of the world, and the state file drifts. smed reads the world and plans against it: go look, compute a diff, show it, then apply it concurrently per host.`,
			`It is a Go library, not a YAML file or a DSL standing in for a real language. Configuration is a Go program that imports smed, and <code>go run .</code> is the dispatcher. Tailscale is the default way to find and reach machines.`,
		},
		Not: []Not{
			{"A state file.", "The machines are the only source of truth."},
			{"An agent.", "Nothing needs to be installed on the targets first."},
		},
	},
	{
		Name:    "molt",
		Tagline: `Builds a folder of plain SQL into a DuckDB database: in dependency order, with tests, in one transaction.`,
		LeansOn: []string{"DuckDB"},
		Install: "go install github.com/" + githubOwner + "/molt/cmd/molt@latest",
		Why: []template.HTML{
			`DuckDB made analytics simple: one engine, one file, no server. The tools around it put a layer between you and the engine. molt removes the layer: it assumes DuckDB and nothing else, and lets DuckDB do every job it can already do.`,
			`Every run is one transaction. A failing model or test rolls everything back, so readers never see half a rebuild or data that failed its tests.`,
		},
		Not: []Not{
			{"Multi-warehouse.", "DuckDB only; that is what keeps it this small."},
			{"A platform.", "No Python runtime, no server, no metadata database."},
		},
	},
}

func toolByName(name string) (Tool, bool) {
	for _, t := range tools {
		if t.Name == name {
			return t, true
		}
	}
	return Tool{}, false
}
