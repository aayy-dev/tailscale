// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux && !ts_omit_androidbin

package cli

// Import androidbin (which also brings in androiddns) so that a static
// linux CLI run on Android can resolve names and verify TLS on its own,
// for commands like netcheck, ping by name, and version --upstream, as
// tailscaled already can.
import _ "tailscale.com/feature/androidbin"
