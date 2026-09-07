package lib

// AssetsVersion is the single cache-busting version for every static asset
// URL (CSS, JS). Bump it whenever a built asset changes so browsers re-fetch
// instead of serving a stale bundle. Centralized here (S4.6): templates render
// {{.AssetsVersion}} from this constant, and the coherence test pins it against
// the npm package version so the three version surfaces cannot drift again.
const AssetsVersion = "0.6.7"
