// Simple Scryfall image fetcher with in-memory cache
// Docs: https://scryfall.com/docs/api/cards/named

const cache = new Map<string, string | null>();

function buildKey(name: string, size: 'normal' | 'small' | 'large') {
  return `${name}|${size}`;
}

function splitFaces(name: string): string[] {
  // Handle split/DFC/meld names like "Fire // Ice"; prefer the first face
  const parts = name.split(/\s*\/\/\s*/);
  if (parts.length > 1) return parts.map((p) => p.trim());
  return [name];
}

export async function fetchScryfallImageByName(name: string, size: 'normal' | 'small' | 'large' = 'normal'): Promise<string | null> {
  const key = buildKey(name, size);
  if (cache.has(key)) return cache.get(key) ?? null;

  const faceCandidates = splitFaces(name);
  const queries: string[] = [];
  // Try exact on first face to avoid 404 on combined meld names
  if (faceCandidates[0]) queries.push(`exact=${encodeURIComponent(faceCandidates[0])}`);
  // Then fuzzy on full name
  queries.push(`fuzzy=${encodeURIComponent(name)}`);
  // Then fuzzy on first face
  if (faceCandidates[0]) queries.push(`fuzzy=${encodeURIComponent(faceCandidates[0])}`);
  // Then fuzzy on second face if exists
  if (faceCandidates[1]) queries.push(`fuzzy=${encodeURIComponent(faceCandidates[1])}`);

  for (const q of queries) {
    try {
      const url = `https://api.scryfall.com/cards/named?${q}`;
      const res = await fetch(url, { mode: 'cors' });
      if (!res.ok) continue; // try next candidate
      const data = await res.json();
      let img: string | undefined;
      if (data.image_uris && data.image_uris[size]) {
        img = data.image_uris[size];
      } else if (Array.isArray(data.card_faces)) {
        // Use first face by default for multi-faced cards
        const firstFace = data.card_faces.find((f: any) => f?.image_uris?.[size]) || data.card_faces[0];
        img = firstFace?.image_uris?.[size];
      }
      const val = img ?? null;
      cache.set(key, val);
      return val;
    } catch {
      // ignore and try next candidate
    }
  }

  cache.set(key, null);
  return null;
}

export function getCachedImage(name: string, size: 'normal' | 'small' | 'large' = 'normal'): string | null | undefined {
  return cache.get(buildKey(name, size));
}
