import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { mount } from '@vue/test-utils';
import { defineComponent } from 'vue';
import { print, type DocumentNode } from 'graphql';
import {
  GAME_UPDATED_SUBSCRIPTION,
  GAMES_QUERY,
  GET_GAME_QUERY,
} from '../src/graphql/queries';
import {
  CREATE_GAME_MUTATION,
  JOIN_GAME_MUTATION,
} from '../src/graphql/mutations';
import { displayNameOf } from '../src/services/displayName';

describe('displayNameOf', () => {
  it.each([
    { label: 'null', DisplayName: null },
    { label: 'undefined', DisplayName: undefined },
    { label: 'empty', DisplayName: '' },
    { label: 'whitespace', DisplayName: ' \t\n ' },
  ])('falls back to Username for a $label display name', ({ DisplayName }) => {
    expect(displayNameOf({ DisplayName, Username: 'Brave Sliver' })).toBe('Brave Sliver');
  });

  it('prefers a non-empty display name', () => {
    expect(displayNameOf({ DisplayName: 'Dylan', Username: 'Brave Sliver' })).toBe('Dylan');
  });

  it('returns an empty string instead of throwing when neither name exists', () => {
    expect(displayNameOf({})).toBe('');
  });

  it('does not treat a duplicate display name as an identity conflict', () => {
    const players = [
      { DisplayName: 'Dylan', Username: 'Brave Sliver' },
      { DisplayName: 'Dylan', Username: 'Clever Wizard' },
    ];
    expect(players.map(displayNameOf)).toEqual(['Dylan', 'Dylan']);
  });
});

describe('display-name data and identity invariants', () => {
  const playerDocuments: [string, DocumentNode][] = [
    ['GAMES_QUERY', GAMES_QUERY],
    ['GET_GAME_QUERY', GET_GAME_QUERY],
    ['GAME_UPDATED_SUBSCRIPTION', GAME_UPDATED_SUBSCRIPTION],
    ['CREATE_GAME_MUTATION', CREATE_GAME_MUTATION],
    ['JOIN_GAME_MUTATION', JOIN_GAME_MUTATION],
  ];

  it.each(playerDocuments)('%s requests DisplayName with Players', (_name, document) => {
    const source = print(document);
    expect(source).toContain('Players');
    expect(source).toMatch(/Players\s*\{[^}]*\bDisplayName\b/s);
  });

  it('continues resolving self by the unique Username when display names match', () => {
    const game = {
      Players: [
        { ID: 'one', Username: 'Brave Sliver', DisplayName: 'Dylan' },
        { ID: 'two', Username: 'Clever Wizard', DisplayName: 'Dylan' },
      ],
    };
    const profile = { Username: 'Clever Wizard', DisplayName: 'Dylan' };

    const self = game.Players.find(player => player.Username === profile.Username);

    expect(self?.ID).toBe('two');
    expect(displayNameOf(self ?? {})).toBe('Dylan');
  });

  it('routes every declared display surface through the shared helper', () => {
    const viewPaths = [
      '../src/components/layout/AppNav.vue',
      '../src/views/ScoreView.vue',
      '../src/views/GamesView.vue',
      '../src/views/BoardView.vue',
      '../src/views/GameAnalysisView.vue',
    ];

    for (const path of viewPaths) {
      const source = readFileSync(new URL(path, import.meta.url), 'utf8');
      expect(source, path).toContain('displayNameOf');
    }
  });

  it('renders markup characters literally through Vue text interpolation', () => {
    const markup = '<img src=x onerror="alert(1)">';
    const wrapper = mount(defineComponent({
      setup: () => ({ displayNameOf, player: { DisplayName: markup, Username: 'Brave Sliver' } }),
      template: '<p>{{ displayNameOf(player) }}</p>',
    }));

    expect(wrapper.text()).toBe(markup);
    expect(wrapper.find('img').exists()).toBe(false);
    expect(wrapper.html()).toContain('&lt;img');
  });
});
