# DiagnosticToolkit local web UI

Mode: Operate. Primary user: a local operator checking their machine's network health. Primary job: start monitoring or run once, then understand the observed results in their network context. Preserve start/stop, run once, live/recent results, run history, per-run logs and summary, editable YAML-backed settings, and environment inspection. The home view prioritizes run controls and current run; when idle it may show previous runs but must not leak old results as current. The Environment page prioritizes interfaces, addresses, gateway, resolvers/upstreams, routes/leases where present, and warnings.

## Direction contract

**THESIS:** Use the visual grammar of technical reference diagrams to make network relationships and check outcomes verifiable; refuse generic stacks of decorative metric cards.

**OWN-WORLD:** Cool drafting-white and graphite surfaces, technical blue for structure and active navigation, amber for warnings, red for failures. Crisp rules and aligned tables; stable humanist UI typography with tabular numerals/monospace only for addresses and measurements. Signature: an accurately labeled, relationship-only network path/topology diagram rendered from actual environment data.

**STORY:** On opening, the operator sees whether monitoring is idle or running and can start monitoring or run once immediately. Results state their expected outcome beside observed status and latency. Environment explains the host's actual network topology; history links each completed run to its logs and summary; settings remain a structured form over the existing YAML config.

**FIRST VIEWPORT:** Shared navigation names Overview, Environment, Logs, History, and Settings. Overview opens with its title, explicit idle/running state, Start/Stop and Run once actions, then a compact topology/path view and readable current results. The topology uses actual gateway, interfaces, and resolver relationships without implying geographic placement. During a run, live results are visible; when idle, prior run cards replace current results.

**FORM:** Assigned direction index 3 of seven grounded directions; user chose the IMPECCABLE'S PICK, “Technical reference diagrams,” on the visual direction round. `seed:97f3a82de00217ce`.

**FINISH:** unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
