import { createLocalVue, mount, shallowMount } from '@vue/test-utils'
import VueRouter from 'vue-router'
import Vuex from 'vuex'
import Board from './../src/components/Board.vue'
import Buefy from 'buefy'
import store from '../src/store'
import router from '../src/router'

// Create a localVue and make it use Vuex
const localVue = createLocalVue()
localVue.use(Vuex)
localVue.use(VueRouter)
localVue.use(Buefy)

describe('board', () => {
    it('renders a boardstate', async () => {
        const wrapper = shallowMount(Board, { router, store, localVue })
        const selfBattlefield = wrapper.find('#selfBattlefield')
        expect(selfBattlefield.exists()).toBe(true)
    })
})