import { createLocalVue, mount, shallowMount } from '@vue/test-utils'
import VueRouter from 'vue-router'
import Vuex from 'vuex'
import App from './../src/App.vue'
import Buefy from 'buefy';
import store from '../src/store';
import { isType } from 'graphql';

describe('edhgo', () => {
    // const wrapper = createLocalVue()

    // test('does a wrapper exist', () => {
    //   expect(wrapper.exists()).toBe(true)
    // })

    // let wrp;
    // const routes = [{ path: '/games/:id', name: 'board' }]
    // const router = new VueRouter({ routes })
    // beforeEach(() => {
    //     const localVue = createLocalVue()
    //     // use our framework, router, and store for tests
    //     localVue.use(VueRouter)
    //     localVue.use(Buefy)
    //     localVue.use(Vuex)

    //     wrp = mount(Board, {
    //         localVue: localVue,
    //         // add on our own store and router 
    //         store,
    //         router,
    //     })
    // })

    // it('should have an App', (t) => {
    //     expect(wrp.exists()).toBe(true)
    // })
    it('renders an app', () => {
        const wrapper = shallowMount(App)
        expect(wrapper.contains('div')).toBe(true)
    })
})