package main

import "html/template"

// githubOwner is the GitHub organization the tools live in.
const githubOwner = "vt100io"

// Tool is one tool on the site. Fields of type template.HTML are written
// here, by me, and may contain markup such as <code>.
type Tool struct {
	Name string
	// The question the tool started from and its answer make up its card
	// on the front page. The principle behind them is on the tool's page.
	Question  string
	Answer    template.HTML
	Principle string
	Early     bool // not there yet
	Tagline   template.HTML
	LeansOn   []string
	Install   string // the command that gets it
	Demo      string // path under /static, optional
	DemoAlt   string
	Why       []template.HTML
	Not       []Not
	Steps     []template.HTML
}

// Not is something a tool deliberately isn't.
type Not struct{ Title, Text string }

func (t Tool) Repo() string { return "https://github.com/" + githubOwner + "/" + t.Name }

var tools = []Tool{
	{
		Name:      "punchcard",
		Question:  "Why does CI need its own build language I can't run on my laptop, and two minutes to deploy an index.html?",
		Answer:    `runs the repo's <code>mise run ci</code> on a machine that remembers.`,
		Principle: "A build is a command I can run myself.",
		Tagline:   `A one-binary CI server that runs the same <code>mise run ci</code> you run on your laptop, on a machine that remembers.`,
		LeansOn:   []string{"mise", "GitHub App"},
		Install:   "go install github.com/" + githubOwner + "/punchcard@latest",
		Demo:      "/static/img/punchcard-tui.gif",
		DemoAlt:   "punchcard tui: a build arrives and streams its log, a newer push replaces a queued one, and a failed build is opened",
		Why: []template.HTML{
			`Hosted CI runners start from nothing on every build: a fresh machine, fresh tool installs, cold caches. Punchcard keeps a workspace per repo on a machine I own. When I push, it fetches, resets to the exact commit and runs the repo's <code>ci</code> task.`,
			`There is no CI language. The pipeline is a mise task in the repo, so if it passes on my laptop it passes in CI, and a CI failure can be debugged on my own machine.`,
			`A GitHub App posts the checks. A repo opts in by having a <code>ci</code> task; there is no list of repos to keep.`,
		},
		Not: []Not{
			{"A sandbox.", "Anyone with write access runs code on the machine. Strangers' pull requests are ignored."},
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
		Name:      "unit",
		Question:  "Why do I put a statically compiled binary into a container?",
		Answer:    `deploys binaries as hardened systemd units over SSH.`,
		Principle: "Nothing between the program and the machine.",
		Tagline:   `Deploys statically compiled binaries as hardened systemd units over SSH. No containers, no daemons, just PID 1.`,
		LeansOn:   []string{"systemd", "SSH"},
		Install:   "go install github.com/" + githubOwner + "/unit/cmd/unit@latest",
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
		Name:      "smed",
		Question:  "Why do I keep an inventory and tags when Tailscale already has both?",
		Answer:    `finds machines in the tailnet, shows the diff, applies it, then keeps watch.`,
		Principle: "Look at the real thing, before you change it and after.",
		Tagline:   `Builds and configures machines from plain Go: it reads each machine, shows a plan, applies it, and then keeps watch over what it built. No interpreter on the target, no state file anywhere.`,
		LeansOn:   []string{"Tailscale", "SSH"},
		Install:   "go get github.com/" + githubOwner + "/smed",
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
		Name:      "molt",
		Question:  "Why does my SQL need a layer around it, when DuckDB already parses, binds and plans it?",
		Answer:    `builds plain SQL into DuckDB, and knows where every column comes from.`,
		Principle: "Let the engine do the work.",
		Early:     true,
		Tagline:   `Builds a folder of plain SQL into a DuckDB database: in dependency order, with tests, in one transaction. It knows where every column comes from, and what a change will touch before it runs.`,
		LeansOn:   []string{"DuckDB"},
		Install:   "go install github.com/" + githubOwner + "/molt/cmd/molt@latest",
		Why: []template.HTML{
			`DuckDB made analytics simple: one engine, one file, no server. The tools around it put a layer between you and the engine. molt removes the layer: it assumes DuckDB and nothing else, and lets DuckDB do every job it can already do.`,
			`Every run is one transaction. A failing model or test rolls everything back, so readers never see half a rebuild or data that failed its tests.`,
			`Column lineage comes from DuckDB too. molt asks DuckDB's binder for the plan of each model and follows every column down to the columns it reads, so <code>SELECT *</code>, <code>GROUP BY ALL</code> and macros resolve exactly as they run. There is no SQL parser in molt and no dialect to guess.`,
			`<code>molt plan</code> compares the project with what is built, column by column: the models you changed, the columns downstream whose data will change, and whether that is breaking. <code>molt run</code> prints the same plan before it builds.`,
			`Incremental models build only the time intervals that are new. The interval is two DuckDB variables, so the model stays plain SQL, and a change that would mix old and new definitions stops the run until you choose a full refresh or forward-only.`,
		},
		Not: []Not{
			{"Multi-warehouse.", "DuckDB only; that is what keeps it this small."},
			{"A platform.", "No Python runtime, no server, no metadata database."},
		},
		Steps: []template.HTML{
			`Write one <code>SELECT</code> per model in <code>models/&lt;schema&gt;/&lt;name&gt;.sql</code>, and settings, secrets and macros in <code>init.sql</code>.`,
			`See what changes: <code>molt plan</code>.`,
			`Build it: <code>molt run</code>.`,
			`Trace a column to its sources: <code>molt lineage &lt;model&gt; &lt;column&gt;</code>.`,
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
