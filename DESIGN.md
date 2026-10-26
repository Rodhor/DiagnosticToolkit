---
name: DiagnosticToolkit
description: Local network diagnostics presented as a technical reference system.
colors:
  primary: "#17679b"
  primary-deep: "#104f79"
  primary-wash: "#e3eff6"
  background: "#eef2f5"
  foreground: "#172536"
  surface: "#ffffff"
  sidebar: "#1b2637"
  sidebar-text: "#dce5ec"
  warning: "#9a6208"
  warning-wash: "#fff1d1"
  success: "#216647"
  success-wash: "#e2f2e9"
  failure: "#a8322c"
  failure-wash: "#fae8e5"
  border: "#cbd5de"
  muted: "#526779"
typography:
  headline:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "clamp(25px, 2.2vw, 32px)"
    fontWeight: 700
    lineHeight: 1.15
    letterSpacing: "-0.035em"
  title:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "16px"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.01em"
  body:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "10px"
    fontWeight: 700
    letterSpacing: "0.065em"
  data:
    fontFamily: "SFMono-Regular, Consolas, Liberation Mono, monospace"
    fontSize: "11px"
    fontWeight: 400
rounded:
  sm: "3px"
  md: "5px"
  lg: "7px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "20px"
  xl: "24px"
  page: "34px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "#ffffff"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "36px"
  button-outline:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "36px"
  button-destructive:
    backgroundColor: "{colors.failure}"
    textColor: "#ffffff"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "36px"
  nav-link:
    textColor: "{colors.sidebar-text}"
    rounded: "4px"
    padding: "0 11px"
    height: "40px"
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    padding: "20px 21px"
  input:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.foreground}"
    rounded: "4px"
    padding: "0 9px"
    height: "36px"
  status-pass:
    backgroundColor: "{colors.success-wash}"
    textColor: "{colors.success}"
    rounded: "{rounded.sm}"
    padding: "3px 7px"
  status-warning:
    backgroundColor: "{colors.warning-wash}"
    textColor: "{colors.warning}"
    rounded: "{rounded.sm}"
    padding: "3px 7px"
  status-failure:
    backgroundColor: "{colors.failure-wash}"
    textColor: "{colors.failure}"
    rounded: "{rounded.sm}"
    padding: "3px 7px"
---

# Design System: DiagnosticToolkit

## Overview

**Creative North Star: "Technical reference diagrams"**

DiagnosticToolkit uses the visual language of a technical reference: cool drafting-white and graphite surfaces, blue rules and links, and compact, aligned data. Its interface is a working instrument for a local operator, not a decorative analytics dashboard. System UI typography keeps labels and explanations familiar; monospace is reserved for addresses and measured values.

The signature is a relationship-only host-to-gateway path built from discovered environment data. Results keep the configured expectation beside the observed status and latency. Monitoring state is explicit, and previous runs remain labeled as history rather than being presented as current results.

**Key Characteristics:**
- Cool light surfaces with a graphite navigation rail and technical blue structure.
- Crisp borders, compact tables, and tabular-looking monospace network values.
- A labeled topology path that reports host relationships, not geographic scale.
- Amber warnings and red failures paired with readable status text.

Source paths: `assets/css/input.css` defines the visual tokens and responsive rules, and `assets/css/output.css` is the stylesheet loaded by the shell. `web/templates/layouts/app.templ` defines shared navigation and landmarks. Page sources are `web/templates/pages/dashboard.templ`, `environment.templ`, `logs.templ`, `history.templ`, `summary.templ`, and `settings.templ`; component sources include `web/templates/components/networkpath.templ`, `results.templ`, `status.templ`, `summary.templ`, and `results-poller.templ`. `web/static/app.js` announces refreshed results accessibly. `PRODUCT.md` records durable product constraints.

## Colors

The palette pairs blue structure with cool neutrals; amber, green, and red distinguish warning, pass, and failure states.

### Primary
- **Technical blue** (`primary`): links, active actions, focus-ring accents, and the emphasized gateway node.
- **Deep technical blue** (`primary-deep`): high-contrast text links.
- **Blue wash** (`primary-wash`): neutral network state surfaces and supporting highlights.

### Neutral
- **Drafting background** (`background`) and **white surface** (`surface`): page canvas and raised-by-boundary content regions.
- **Graphite rail** (`sidebar`) with **pale rail text** (`sidebar-text`): persistent desktop navigation.
- **Ink** (`foreground`) and **muted slate** (`muted`): primary and supporting text.
- **Rule gray** (`border`): table dividers, field strokes, and panel boundaries.

### Named Rules
**The Evidence Color Rule.** Status colors reinforce explicit words such as PASS, WARN, and FAIL; color is not the only status cue.

## Typography

**Display Font:** No separate display face; headings use the system UI stack.  
**Body Font:** `system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`.  
**Label/Mono Font:** `SFMono-Regular, Consolas, "Liberation Mono", monospace` for IP addresses, targets, run references, and measurements.

**Character:** Compact, stable system typography favors scanability over expressive display styling. Monospace makes machine values visually distinct without extending to prose.

### Hierarchy
- **Headline** (700, responsive 25–32px, 1.15): page titles, with tight negative tracking.
- **Title** (600, 16px, 1.3): section headings; a few prominent idle-state headings use 20px.
- **Body** (400, 14px, 1.5): page introductions and explanatory copy; dense tables and controls step down to 12px.
- **Label** (700, 10px, 0.065em, uppercase where used): compact table headings and section labels.
- **Data** (400, 11px, normal tracking): monospace addresses and measurements; wrap long values where the table allows it.

## Layout

At desktop widths the app uses a 232px fixed navigation rail and a flexible main column. The main content is centered up to 1500px, with horizontal padding that scales from 20px to 44px and 34px top padding. A 54px white top bar separates the app context from page content. Content sections use a compact 22px vertical rhythm, white surfaces, crisp one-pixel rules, and roughly 20px internal padding.

At 1000px and below, the rail narrows to 190px and four-column settings groups reduce to two columns. At 760px and below, the shell becomes one column: navigation moves above the content into a horizontally scrollable row, the top bar shortens and hides its secondary note, and page headings and controls stack. Settings grids become one column; the network path stacks vertically with its connector rotated; three-column environment summaries stack. Wide data tables scroll horizontally instead of compressing their values. The result list has its own bounded vertical scroll region.

The desktop topology is a simple host → default gateway relationship with the interface/default route label on the connector. It is not a geographic map and it only displays discovered data. The Overview prioritizes run controls and network context; Environment uses grouped details and tables; Logs, History, Summary, and Settings retain dense task-specific structures.

## Elevation & Depth

Depth is mostly structural, not shadow-driven: white panels sit on a cool gray canvas and use thin rules to define boundaries. The shared templui buttons carry a subtle one-pixel/two-pixel shadow; content panels and topology nodes remain flat. Blue-tinted fills distinguish selected or emphasized information from plain surfaces.

### Shadow Vocabulary
- **Control shadow** (`0 1px 2px 0 rgb(0 0 0 / 0.05)`): subtle depth on templui action buttons only.

## Shapes

The form language is crisp and lightly softened rather than pill-shaped: common panels use 5px corners, small status tags use 3px, and controls/navigation use 4–5px. Borders are fine and cool gray. Tables use aligned rows and horizontal dividers; topology nodes are rectangular outlined boxes joined by a thin directional line.

## Components

### Buttons
- **Character:** compact, direct actions with distinct run, stop, and secondary emphasis.
- **Shape:** gently rounded (5px); default height is 36px with 8px by 16px padding.
- **Primary:** technical blue fill with white text; used by the templui default variant, including Start monitoring and Save and reload.
- **Destructive:** red fill with white text for Stop monitoring.
- **Outline:** background-colored with a border and subtle shadow for Run once; hover uses the secondary accent surface.
- **Hover / Focus:** variant hover colors are supplied by templui. Buttons expose a visible themed focus ring; global focus-visible styling also uses a 3px amber outline offset by 2px where not overridden.
- **Disabled:** templui buttons suppress pointer events and reduce opacity.

### Chips
- **Style:** compact, lightly rounded labels with state-colored text and wash backgrounds; pass, warning, failure, and unknown have distinct treatments.
- **State:** textual status remains present alongside color. Monitoring state uses explicit Idle/Monitoring text; the live state adds a green indicator with a restrained pulse.

### Cards / Containers
- **Corner Style:** 5px.
- **Background:** white on the cool gray page canvas; topology and idle panels use pale blue-gray fills.
- **Shadow Strategy:** flat, boundary-led surfaces; see Elevation & Depth.
- **Border:** one-pixel cool gray rule, with a warm border for warning sections.
- **Internal Padding:** generally 20–21px, reduced to 15px on narrow screens.

### Inputs / Fields
- **Style:** white fill, one-pixel gray stroke, 4px corners, 36px height, and 9px horizontal padding. Labels and help text are adjacent to their controls.
- **Focus:** blue border and a soft blue outline on fields; disabled query values use a gray fill and not-allowed cursor.
- **Error / Disabled:** save/load errors use a red notice with `role="alert"`; saved confirmation uses a green notice with `role="status"`.

### Navigation
- **Style:** dark graphite desktop rail, 13px links, 40px minimum link height, and 4px corners. The selected link uses a lighter graphite fill and `aria-current="page"`; hover lightens the row. A small blue marker accompanies the selected item.
- **Labels:** Overview, Environment, Live logs, Run history, and Settings.
- **Mobile:** the rail becomes a compact header with a horizontally scrollable navigation row; the local-session footer and active marker are hidden.

### Network Path
The signature component is a factual, labeled host-to-default-gateway diagram. Nodes show the discovered host/interface count and gateway, while the connector names the route interface when available. The Overview also lists system resolvers separately. Missing gateway data is shown as unavailable rather than inferred; the Environment page explains that it was not detected.

### Data Tables
Results and environment tables use compact 10px uppercase headers, 12px body text, horizontal rules, and a subtle row hover. Addresses, targets, and measurements use monospace. Tables scroll in their own container when the available width is insufficient.

### Accessibility and States
The shared shell uses navigation, header, and main landmarks; the current page is exposed with `aria-current`. Sections have heading associations, fields have labels or accessible names, and notices/statuses use live-region roles. Result polling refreshes every two seconds while its table is excluded from noisy live announcements; a separate polite, atomic status region announces observation updates. Keyboard focus remains visible. Pass/warn/fail/unknown and interface states have words as well as color. Empty and missing-data cases use explicit copy rather than fabricated topology.

## Do's and Don'ts

### Do:
- **Do** use the discovered host, gateway, and route relationship for the topology diagram; retain its labels and directional connector.
- **Do** keep expected outcomes visible beside observed result status and latency.
- **Do** distinguish the current monitoring state from prior runs and label missing host data as unavailable.
- **Do** keep wide tables scrollable and allow long network values to wrap where the layout permits.

### Don't:
- **Don't** imply geographic scale or add network nodes that were not discovered.
- **Don't** communicate a result state with color alone.
- **Don't** present previous run results as the current run when monitoring is idle.
