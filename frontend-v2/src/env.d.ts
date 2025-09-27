/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_GRAPHQL_HTTP?: string;
	readonly VITE_GRAPHQL_WS?: string;
	readonly VITE_DEV_PORT?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}

declare const __APP_VERSION__: string;
