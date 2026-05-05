# Multi-Agent Namespace Example

This example demonstrates how to bypass the 512 Tool limitation by using `namespace` to create specialized agents.
A central Orchestrator (Client) communicates with a Router Agent (`ask_web_agent`), which then delegates the task to a specialized Web Agent (`web@httpd_log`) using the `web` namespace.

## How to run

Start Registry:
```bash
go run registry.go
```

Start Web Agent:
```bash
go run web-agent.go
```

Start Router Agent:
```bash
go run router-agent.go
```

Run Client:
```bash
go run client.go
```
