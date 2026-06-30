// Static SPA: no server-side rendering and no server runtime. The game relies on
// browser-only APIs (Canvas, WebSocket, window) and talks to the Go server purely
// over /ws, so we prerender a single shell and render everything client-side.
export const prerender = true;
export const ssr = false;
