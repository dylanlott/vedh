import { createLocalVue, mount } from '@vue/test-utils'
import VueRouter from 'vue-router'
import { Boardstates, Users, Games } from '../src/store/'
import Vuex from 'vuex'
import Board from '../src/components/Board.vue'
import router from '../src/router'
import Buefy from 'buefy'

describe('board', () => {
  let state;
  let actions;
  let mutations;
  let store;
  let localVue;

  beforeEach(() => {
    state = {
      boardstates: {},
      self: {},
    }
    actions = {
      // mock `mutateBoardState` so that we don't need to make an API call.
      mutateBoardState: function ({ commit }, payload) {
        console.log('commit hit: ', [payload])
        commit('updateBoardStates', [payload])
      },
      subAllBoardstates: jest.fn(),
      move: Boardstates.actions.move,
      error: jest.fn()
    }
    mutations = Boardstates.mutations
    store = new Vuex.Store({
      modules: {
        Boardstates: {
          state,
          mutations,
          actions,
        },
        Users,
        Games,
      }
    })
    localVue = createLocalVue()
    localVue.use(Vuex)
    localVue.use(VueRouter)
    localVue.use(Buefy)
  })

  it('renders a boardstate', async () => {
    const wrapper = mount(Board, { router, store, localVue })
    const selfBattlefield = wrapper.find('#selfBattlefield')
    expect(selfBattlefield.exists()).toBe(true)
  })

  it('moves target card from source to destination', async () => {
    const wrapper = mount(Board, { localVue, router, store })
    const battlefield = wrapper.find('#selfBattlefield')
    expect(battlefield.exists()).toBe(true)
    state.boardstates['shakezula'] = {
      User: {
        ID: 'shakezula'
      },
      Hand: [
        { ID: 'abc123', Name: 'Kykar, Wind\'s Fury' },
        { ID: 'def456', Name: 'Rough // Tumble' },
        { ID: 'ghi789', Name: 'Plains' },
        { ID: 'jkl123', Name: 'Polluted Delta' },
      ],
      Field: [],
      Exile: [],
      Graveyard: [],
    }
    expect(state.boardstates['shakezula'].User.ID).toEqual('shakezula')
    // dispatch a move action and check the board for states
    store.dispatch('move', {
      userID: 'shakezula',
      src: 'Hand',
      dest: 'Field',
      target: 'ghi789'
    })
    expect(state.boardstates['shakezula'].Field)
      .toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            ID: 'ghi789', Name: 'Plains',
          })
        ]))
    expect(state.boardstates['shakezula'].Hand)
      .toEqual(expect.arrayContaining([
        expect.objectContaining({
          ID: 'abc123', Name: 'Kykar, Wind\'s Fury',
        }),
        expect.objectContaining({
          ID: 'jkl123', Name: 'Polluted Delta',
        }),
        expect.objectContaining({
          ID: 'def456', Name: 'Rough // Tumble',
        }),
      ]))
  })
})