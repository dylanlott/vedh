# vEDH – Frontend

A ground-up modernization of the original EDH Go frontend, rebranded as vEDH, using Vue 3, Vite, Pinia, and Apollo Client. The goal is to preserve the existing product features—authentication, lobby management, board state control, and rich card interactions—while adopting a contemporary stack and a more maintainable architecture.

## Getting started

```bash
# install dependencies
npm install

# start the dev server
npm run dev

# run a production build
npm run build

# type-check the project
npm run type-check
```

By default, the GraphQL API is expected at `http://localhost:8080/graphql`. Override the endpoints by adding a `.env` file with the following settings:

```bash
VITE_GRAPHQL_HTTP=http://localhost:8080/graphql
VITE_GRAPHQL_WS=ws://localhost:8080/graphql
```

## Current progress

- ✅ Vite + Vue 3 application scaffold
- ✅ Apollo client (HTTP + websocket) wired to the existing GraphQL API
- ✅ Pinia stores for authentication and game state
- ✅ Initial screens: landing page, login, signup, lobby, board, join flow, score, card view
- ✅ Minimal create-game modal and live game subscription plumbing
- 🔜 Drag-and-drop commander board with zone management
- 🔜 Automated GraphQL code generation
- 🔜 Comprehensive UI polish & accessibility pass

## Architecture highlights

- **Vue 3 + `<script setup>`** for concise, type-safe components.
- **Pinia** modules mirror the legacy Vuex structure (users, games, cards) while embracing Composition API patterns.
- **Apollo Client 3** powers queries, mutations, and subscriptions with auth-aware links.
- **Modular routing** with per-route guards that reuse the Pinia auth state.
- **Sass utilities** and a dark, glassmorphism-inspired baseline styling to keep gameplay readable.

Refer to the project roadmap in the root README for the broader migration plan.
