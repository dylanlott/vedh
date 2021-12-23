import { createLocalVue, mount, shallowMount } from '@vue/test-utils'
import VueRouter from 'vue-router'
import Vuex from 'vuex'
import App from './../src/App.vue'
import Buefy from 'buefy';
import store from '../src/store';
import { isType } from 'graphql';

describe('edhgo', () => {
    it('renders an app', () => {
        const wrapper = shallowMount(App)
        expect(wrapper.contains('div')).toBe(true)
    })
})