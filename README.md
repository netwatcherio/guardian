# NetWatcher.io Guardian

NetWatcher.io Guardian is a Go application that serves as the backend for managing NetWatcher agents, probes, and probe data. It also handles WebSocket communication to receive data from these agents.

## Naming Conventions

- **Agents:** Agents in the NetWatcher.io Guardian application are software components installed on various networks or controlled by Managed Service Providers (MSPs) at different sites. These agents act as the "eyes and ears" of the network monitoring system, responsible for collecting data and monitoring network conditions.

- **Probes:** Probes are individual network checks or monitoring tasks initiated and executed by the agents. These tasks can include running network diagnostics such as MTR (My Traceroute), ping tests, or other network-related checks. Probes collect specific data points or measurements about network performance and reliability.

- **Probe Data:** Probe data represents the results and measurements obtained from the various probes executed by the agents. This data includes information about network latency, packet loss, round-trip times, route information (in the case of MTR), and other relevant metrics. Probe data is sent back to the Guardian backend for analysis, storage, and presentation.

## Prerequisites

Before you begin, ensure you have met the following requirements:

- [Go](https://golang.org/) installed on your system.
- Proper configuration for database connections and external services.

## Installation

1. Clone this repository:

   ```bash
   git clone https://github.com/yourusername/netwatcher-guardian.git
   cd netwatcher-guardian
