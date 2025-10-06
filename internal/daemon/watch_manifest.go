package daemon

// watch_manifest.go parses and reconciles persisted manifests. Collaborates with
// pkg/config and internal/state to reconcile desired vs actual watch targets.

// TODO: Implement the manifest reconciliation logic.
// This should include:
// - A function to load the manifest from disk using the state.ManifestStore.
// - Logic to compare the loaded manifest with the current watcher configuration.
// - Functions to add or remove watched directories from a running watcher.Controller
//   based on changes in the manifest.
