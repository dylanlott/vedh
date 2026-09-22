export type DeckBracket = 1 | 2 | 3 | 4 | 5;

export interface DefaultDeck {
  id: string;
  bracket: DeckBracket;
  bracketName: string;
  name: string;
  commander: string;
  description: string;
  updatedAt: string;
  sourceLabel: string;
  sourceUrl: string;
  entries: readonly (readonly [number, string])[];
}

const officialBracketSource = 'https://magic.wizards.com/en/formats/commander';

function deck(id: string, bracket: DeckBracket, bracketName: string, name: string, commander: string, description: string, cards: string[], lands: readonly (readonly [number, string])[]): DefaultDeck {
  return {
    id, bracket, bracketName, name, commander, description,
    updatedAt: '2026-09-21', sourceLabel: 'vEDH curated · Commander Brackets', sourceUrl: officialBracketSource,
    entries: [[1, commander], ...cards.map(card => [1, card] as const), ...lands],
  };
}

// App-owned snapshots keep game creation reliable and make refresh dates reviewable.
export const defaultDecks: readonly DefaultDeck[] = [
  deck('b1-pantlaza-dinosaur-showcase', 1, 'Exhibition', 'Pantlaza Dinosaur Showcase', 'Pantlaza, Sun-Favored', 'Theme-first Dinosaurs with splashy discovers and a relaxed mana base.', [
    'Ancient Imperiosaur', 'Apex Altisaur', 'Atzocan Seer', "Burning Sun's Avatar", 'Colossal Dreadmaw', 'Curious Altisaur', 'Earthshaker Dreadmaw', 'Etali, Primal Storm', 'Etali, Primal Conqueror', 'Ghalta and Mavren', 'Ghalta, Primal Hunger', "Gishath, Sun's Avatar", 'Huatli, Poet of Unity', 'Itzquinth, Firstborn of Gishath', "Kinjalli's Sunwing", 'Marauding Raptor', 'Otepec Huntmaster', "Palani's Hatcher", 'Quartzwood Crasher', 'Raging Regisaur', 'Ranging Raptors', 'Ravenous Tyrannosaurus', 'Regisaur Alpha', 'Ripjaw Raptor', 'Runic Armasaur', 'Scion of Calamity', 'Sunfrill Imitator', 'Temple Altisaur', 'Thrashing Brontodon', 'Topiary Stomper', 'Trumpeting Carnosaur', 'Wayward Swordtooth', 'Zacama, Primal Calamity', 'Arcane Signet', "Commander's Sphere", 'Cultivate', 'Dinosaur Stampede', 'Explore', "Kodama's Reach", 'Migration Path', 'Naya Charm', 'Return of the Wildspeaker', 'Sol Ring', 'Thunderherd Migration', 'Unnatural Growth', 'Welcome to . . .', 'Cabaretti Courtyard', 'Canopy Vista', 'Cinder Glade', 'Command Tower', 'Exotic Orchard', 'Fortified Village', 'Game Trail', 'Jungle Shrine', 'Krosan Verge', 'Myriad Landscape', 'Path of Ancestry', "Rogue's Passage", 'Sungrass Prairie', 'Temple of Abandon', 'Temple of Plenty', 'Temple of Triumph',
  ], [[13, 'Forest'], [11, 'Mountain'], [13, 'Plains']]),
  deck('b2-bello-animated-enchantments', 2, 'Core', 'Bello Animated Enchantments', 'Bello, Bard of the Brambles', 'A straightforward enchantment-and-artifact precon-style game plan.', [
    'Alexios, Deimos of Kosmos', 'Brightcap Badger', "Burning Sun's Avatar", 'Calamity, Galloping Inferno', 'Champion of Lambholt', 'Enduring Courage', 'Etali, Primal Storm', "Garruk's Packleader", 'Goreclaw, Terror of Qal Sisma', 'Grothama, All-Devouring', 'Gruul Ragebeast', 'Kogla, the Titan Ape', 'Loyal Guardian', 'Ohran Frostfang', 'Rhonas the Indomitable', 'Rootweaver Druid', 'Sakura-Tribe Elder', 'Sunspine Lynx', 'Toski, Bearer of Secrets', 'Wildsear, Scouring Maw', 'Arcane Signet', "Bootleggers' Stash", 'Gruul Signet', 'Hedron Archive', 'Mind Stone', 'Sol Ring', 'Talisman of Impulse', 'Thran Dynamo', 'Unstable Obelisk', "Berserkers' Onslaught", 'Court of Ire', 'Elemental Bond', "Garruk's Uprising", 'Greater Good', 'Outpost Siege', 'Primeval Bounty', 'Rain of Riches', "Sunbird's Invocation", 'Warstorm Surge', 'Wild Growth', 'Beast Within', 'Blasphemous Act', 'Chaos Warp', 'Cultivate', 'Decimate', 'Explore', 'Harmonize', "Kodama's Reach", 'Return of the Wildspeaker', 'Abrade', 'Cinder Glade', 'Command Tower', 'Exotic Orchard', 'Game Trail', 'Gruul Turf', 'Karplusan Forest', 'Kessig Wolf Run', 'Mossfire Valley', 'Mosswort Bridge', 'Myriad Landscape', 'Path of Ancestry', 'Rootbound Crag', 'Temple of Abandon',
  ], [[18, 'Forest'], [18, 'Mountain']]),
  deck('b3-meren-graveyard-value', 3, 'Upgraded', 'Meren Graveyard Value', 'Meren of Clan Nel Toth', 'Upgraded sacrifice value with resilient recursion and efficient answers.', [
    'Apprentice Necromancer', 'Birds of Paradise', 'Blood Artist', 'Carrion Feeder', 'Caustic Caterpillar', 'Dauthi Voidwalker', 'Deathrite Shaman', 'Elves of Deep Shadow', 'Elvish Mystic', 'Eternal Witness', 'Fleshbag Marauder', 'Foundation Breaker', 'Gray Merchant of Asphodel', 'Haywire Mite', 'Llanowar Elves', 'Massacre Wurm', 'Mikaeus, the Unhallowed', 'Plaguecrafter', 'Protean Hulk', 'Reclamation Sage', 'Sakura-Tribe Elder', 'Shriekmaw', 'Sidisi, Undead Vizier', 'Spore Frog', "Stitcher's Supplier", 'Viscera Seer', 'Woe Strider', 'Zulaport Cutthroat', 'Arcane Signet', "Ashnod's Altar", 'Skullclamp', 'Sol Ring', 'Birthing Pod', 'Animate Dead', 'Bastion of Remembrance', 'Grave Pact', 'Necromancy', 'Phyrexian Arena', 'Survival of the Fittest', "Assassin's Trophy", 'Beast Within', 'Buried Alive', 'Culling Ritual', 'Demonic Tutor', 'Diabolic Intent', 'Eldritch Evolution', 'Entomb', 'Finale of Devastation', 'Living Death', "Nature's Claim", 'Reanimate', 'Victimize', 'Bala Ged Recovery', 'Bojuka Bog', 'Command Tower', 'Deathcap Glade', 'Exotic Orchard', 'High Market', 'Llanowar Wastes', 'Myriad Landscape', 'Overgrown Tomb', 'Path of Ancestry', 'Phyrexian Tower', 'Temple of Malady', 'Twilight Mire', 'Undergrowth Stadium',
  ], [[17, 'Forest'], [16, 'Swamp']]),
  deck('b4-yuriko-optimized-tempo', 4, 'Optimized', 'Yuriko Optimized Tempo', "Yuriko, the Tiger's Shadow", 'Fast, lethal Ninja tempo with compact interaction and top-deck setup.', [
    'Changeling Outcast', 'Faerie Seer', 'Fourth Bridge Prowler', 'Gingerbrute', 'Hope of Ghirapur', 'Memnite', 'Mist-Cloaked Herald', 'Moon-Circuit Hacker', 'Network Disruptor', 'Ornithopter', 'Prosperous Thief', "Sakashima's Student", 'Silver-Fur Master', 'Spectral Sailor', 'Thousand-Faced Shadow', 'Triton Shorestalker', 'Universal Automaton', 'Ingenious Infiltrator', 'Mist-Syndicate Naga', "Nashi, Moon Sage's Scion", 'Walker of Secret Ways', 'Arcane Signet', "Sensei's Divining Top", 'Sol Ring', 'Talisman of Dominance', 'Mystic Remora', 'Rhystic Study', 'Bident of Thassa', 'Brainstorm', 'Consider', 'Counterspell', 'Cyclonic Rift', 'Demonic Tutor', 'Dig Through Time', 'Drown in the Loch', 'Fierce Guardianship', 'Force of Negation', 'Force of Will', "Lim-Dul's Vault", 'Mana Drain', 'Mystical Tutor', 'Ponder', 'Preordain', 'Snuff Out', 'Spell Pierce', 'Temporal Trespass', 'Treasure Cruise', 'Vampiric Tutor', 'Wash Away', 'Watery Grave', 'Command Tower', 'Darkslick Shores', 'Drowned Catacomb', 'Exotic Orchard', 'Fabled Passage', 'Flooded Strand', 'Marsh Flats', 'Misty Rainforest', 'Morphic Pool', 'Otawara, Soaring City', 'Polluted Delta', 'Prismatic Vista', 'Scalding Tarn', 'Shipwreck Marsh', 'Underground River', 'Verdant Catacombs', 'Bloodstained Mire',
  ], [[16, 'Island'], [16, 'Swamp']]),
  deck('b5-kinnan-cedh', 5, 'cEDH', 'Kinnan Basalt cEDH', 'Kinnan, Bonder Prodigy', 'Tournament-minded Simic turbo-combo built around Kinnan and Basalt Monolith.', [
    'Birds of Paradise', 'Bloom Tender', 'Delighted Halfling', 'Elvish Mystic', 'Elvish Spirit Guide', 'Fyndhorn Elves', 'Llanowar Elves', 'Mox Amber', 'Ornithopter of Paradise', 'Phyrexian Metamorph', 'Seedborn Muse', 'Spellseeker', "Thassa's Oracle", 'Thrasios, Triton Hero', 'Tidespout Tyrant', 'Treasure Mage', 'Void Winnower', 'Arcane Signet', 'Basalt Monolith', 'Chrome Mox', 'Fellwar Stone', 'Grim Monolith', 'Jeweled Lotus', 'Lotus Petal', 'Mana Vault', 'Mox Diamond', 'Sol Ring', 'Talisman of Curiosity', 'The One Ring', 'Carpet of Flowers', 'Mystic Remora', 'Rhystic Study', 'Sylvan Library', 'Worldly Tutor', 'Brainstorm', 'Chain of Vapor', 'Chord of Calling', 'Cyclonic Rift', 'Dispel', 'Eldritch Evolution', 'Fierce Guardianship', 'Finale of Devastation', 'Flusterstorm', 'Force of Negation', 'Force of Vigor', 'Force of Will', 'Mental Misstep', 'Mindbreak Trap', 'Muddle the Mixture', 'Mystical Tutor', "Nature's Claim", 'Neoform', 'Pact of Negation', 'Pongify', 'Resculpt', 'Swan Song', 'Transmute Artifact', 'Veil of Summer', 'Ancient Tomb', 'Boseiju, Who Endures', 'Botanical Sanctum', 'Breeding Pool', 'Cephalid Coliseum', 'City of Brass', 'Command Tower', 'Exotic Orchard', 'Flooded Strand', 'Gemstone Caverns', 'Mana Confluence', 'Misty Rainforest', 'Otawara, Soaring City', 'Polluted Delta', 'Rejuvenating Springs', 'Scalding Tarn', 'Tropical Island', 'Verdant Catacombs', 'Windswept Heath', 'Wooded Foothills',
  ], [[10, 'Forest'], [11, 'Island']]),
];

function csvCell(value: string): string {
  return /[",\n]/.test(value) ? `"${value.replace(/"/g, '""')}"` : value;
}

export function decklistToCsv(deck: DefaultDeck): string {
  return deck.entries.map(([quantity, name]) => `${quantity},${csvCell(name)}`).join('\n');
}

export function deckCardCount(deck: DefaultDeck): number {
  return deck.entries.reduce((total, [quantity]) => total + quantity, 0);
}

export function commanderPickFor(deck: DefaultDeck) {
  return { ID: `default:${deck.id}:commander`, Name: deck.commander };
}
