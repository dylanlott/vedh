import Vuex from 'vuex'
import Vue from 'vue'
import gql from 'graphql-tag'
import Cookies from 'js-cookie'
import { uuid } from '@/uuid'
import { ToastProgrammatic as Toast } from 'buefy'
import api from '@/gqlclient'
import router from '@/router'
import {
    // auth
    login,
    signup,
    // cards
    cardQuery,
    commanderQuery,
    // games
    getGameQuery,
    gameUpdatedSubscription,
    updateGame,
} from '@/gqlQueries'

Vue.use(Vuex)

const ls = window.localStorage

export const Games = {
    namespaced: true,
    state: {
        game: {
            ID: "",
            Turn: {
                Player: "",
                Phase: "",
                Number: 0
            },
            Players: [],
        },
        error: undefined,
        loading: false,
    },
    mutations: {
        error(state, err) {
            state.error = err
            Toast.open({
                message: `Game error: ${state.error}`,
                duration: 3000,
                position: "is-bottom",
                type: "is-danger",
            })
        },
        updateGame(state, game) {
            state.game = game
        },
        updateTurn(state, turn) {
            state.game.Turn = turn
        },
        gameFailure(state, error) {
            state.error = error
        },
        updateLoading (state, loading) {
            state.loading = loading
        },
    },
    actions: {
        getGame({ commit }, { gameID }) {
            commit('updateLoading', true)
            api.query({
                query: getGameQuery,
                variables: {
                    gameID: gameID 
                }
            }).then((resp) => {
                commit('updateLoading', false)
                commit('updateGame', resp.data.getGame)
                return resp.data.getGame
            }).catch((err) => {
                commit('updateLoading', false)
                console.error('apollo vuex failed to get game: ', err)
                commit('gameFailure', err)
                return err
            })
        },
        subscribeToGame({ commit }, { gameID, userID }) {
            const sub = api.subscribe({
                query: gameUpdatedSubscription,
                variables: {
                    gameID: gameID,
                    userID: userID,
                }
            })
            sub.subscribe({
                next(data) { 
                    console.log("### subscription event recvd: ", data)
                    commit('updateGame', data.data.gameUpdated) 
                },
                error(err) {
                    console.error('vuex error: subscribeToGame: game subscription error: ', err)
                    commit('error', err)
                }
            })
        },
        joinGame({ commit }, payload) {
            commit('updateLoading', true)
            return new Promise((resolve, reject) => {
                api.mutate({
                    mutation: gql`mutation ($InputJoinGame: InputJoinGame) {
                        joinGame(input: $InputJoinGame) {
                            ID
                        }
                    }`,
                    variables: {
                        InputJoinGame: payload.inputGame,
                    }
                })
                .then((res) => {
                    commit('updateGame', res.data.joinGame)
                    commit('updateLoading', false)
                    router.push({ path: `/games/${res.data.joinGame.ID}` })
                    return resolve(res)
                })
                .catch((err) => {
                    commit('error', 'error joining game')
                    commit('updateLoading', false)
                    console.error('error joining game: ', err)
                    return reject(err)
                })
            })
        },
        createGame({ commit }, payload) {
            commit('updateLoading', true)
            return new Promise((resolve, reject) => {
                api.mutate({
                    mutation: gql`mutation ($inputGame: InputCreateGame!) {
                    createGame(input: $inputGame){
                        ID	
                        CreatedAt 
                        Turn {
                            Number
                            Player
                            Phase
                        }
                        Players {
                            Username
                            ID
                        }
                    }
                    }`,
                    variables: {
                        inputGame: payload
                    }
                })
                .then((res) => {
                    commit('updateLoading', false)
                    commit('updateGame', res.data.createGame)
                    router.push({ path: `/games/${res.data.createGame.ID}` })
                    return resolve(res)
                })
                .catch((err) => {
                    commit('error', 'error creating game')
                    commit('updateLoading', false)
                    console.error('error createGame: ', err)
                    return reject(err)
                })
            })
        },
        // sync copies the current game state and then attempts to save it.
        sync({ commit }, payload) {
            return api.mutate({
                mutation: updateGame,
                variables: {
                    input: payload,
                }
            })
            .then((data) => {
                commit('updateGame', data.data.updateGame)
                return data
            })
            .catch((err) => {
                console.error('updateGame failed: ', err)
                return err
            })
        },
    },
}

export const Users = {
    namespaced: true,
    state: {
        User: {
            Username: Cookies.get("username") || ls.getItem("username"),
            ID: Cookies.get("userID") || ls.getItem("userID") || uuid(),
            Token: Cookies.get("token") || ls.getItem("token")
        },
        loading: false,
        error: undefined,
    },
    getters: {
        authenticated: state => {
            return !!state.User.Token
        },
    },
    mutations: {
        setUser(state, payload) {
            state.User.ID = payload.ID
            state.User.Username = payload.Username
            state.User.Token = payload.Token
            Cookies.set("userID", payload.ID)
            Cookies.set("username", payload.Username)
            Cookies.set("token", payload.Token)
            Cookies.set("user_info", JSON.stringify(payload))
        },
        logout(state) {
            // set back to default state
            state.User = {Username: "", Token: "", ID: ""}
            // and then remove all cookies
            Cookies.remove("username")
            Cookies.remove("userID")
            Cookies.remove("token")
            Cookies.remove("user_info")
        },
        loading(state, bool) {
            state.loading = bool
        },
        error(state, message) {
            state.error = message
            Toast.open({
                message: `User error: ${state.error}`,
                duration: 3000,
                position: "is-bottom",
                type: "is-danger",
            })
        }
    },
    actions: {
        login({ commit }, payload) {
            commit('loading', true)
            api.mutate({
                mutation: login,
                variables: {
                    username: payload.username,
                    password: payload.password,
                }
            })
            .then((data) => {
                commit('setUser', data.data.login)
                router.push({ path: '/games' });
                return data
            })
            .catch((err) => {
                console.error('login error: ', err)
                commit('error', 'failed to login')
                return err
            })
        },
        logout({ commit }) {
            commit('logout')
            router.push({ path: '/' })
        },
        signup({ commit, dispatch }, payload) {
            commit('loading', true)
            api.mutate({
                mutation: signup,
                variables: {
                    username: payload.username,
                    password: payload.password,
                }
            })
            .then((resp) => {
                dispatch('login', payload) 
                .catch((err) => commit('error', 'failed to login'))
            })
            .catch((err) => {
                commit('error', 'failed to signup')
            })
        }
    }
}

export const Cards = {
    namespaced: true,
    state: {
        list: [],
    },
    mutations: {
        setList: function(state, payload) {
            state.list = payload
        }
    },
    actions: {
        fetchCard({ commit }, name) {
            return new Promise((resolve, reject) => {
                return api.query({
                    query: cardQuery,
                    variables: {
                        name: name,
                    }
                })
                .then((resp) => {
                    resolve(resp.data.card)
                    commit('setList', resp.data.card)
                })
                .catch((err) => reject(err))
            })
        },
        searchByName({ commit }, name) {
            api.query({
                query: commanderQuery,
                variables: {
                    name: name,
                },
            })
            .then((resp) => {
                const cards = resp.data.search.map((item) => {
                    const card = Object.assign({}, item)
                    console.log("searchByName found card: ", card)
                    return card
                })
                commit('setList', cards)
                return cards
            })
            .catch((err) => {
                console.error('error searching cards by name: ', err);
                commit('error', err)
                return err;
            })
        }
    },
}

export const store = new Vuex.Store({
  modules: {
    Cards,
    Games,
    Users,
  }
})
