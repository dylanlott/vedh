# vEDH Auth Storage Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace vEDH's current `localStorage` bearer-token persistence with a production-safer auth model.

**Architecture:** Move away from long-lived bearer tokens in browser storage. Preferred end state is an `HttpOnly` cookie-backed session issued by the backend. If that is too large for the first slice, use short-lived in-memory access tokens with a refresh path as an interim step.

**Tech Stack:** Vue 3, Pinia, Apollo GraphQL, Go GraphQL backend, browser cookies.

---

### Current state

Confirmed current behavior:
- `vedh/app/src/stores/auth.ts` loads and stores `{ ID, Username, Token }` in `localStorage`
- Apollo attaches that bearer token to requests

This is workable for velocity but not a good long-term production default.

### Recommended migration order

1. Add backend session-cookie support
2. Update login/signup responses and frontend auth bootstrap to use cookies
3. Remove persistent bearer-token storage from `localStorage`
4. Add logout/session-expiry handling
5. Add regression coverage for browser auth bootstrapping

### Minimum acceptance criteria

- No long-lived auth token persisted in `localStorage`
- Refreshing the app preserves a valid session via cookie
- Expired/invalid session redirects cleanly to login
- Apollo requests continue working without manual token plumbing in local storage

### Notes

This should be treated as the next productionization milestone after the logging/metrics/docs cleanup landed on 2026-05-15.
