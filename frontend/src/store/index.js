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
    gameQuery,
    gameUpdateQuery,
    updateGame,
    // boardstates
    boardstates,
    boardstateSubscription,
    updateBoardStateQuery,
} from '@/gqlQueries'

Vue.use(Vuex)

const ls = window.localStorage

export const Boardstates = {
    // BoardStates should only ever be updated by push from the server, we 
    // don't normally want to update these ourselves. 
    state: {
        // boardstates holds an object with all player boardstates
        // keyed by the player's UUID (ID)
        boardstates: {},
        // we load our personal state into self but we still don't want to 
        // violate our one-direction data flow principle
        self: {},
        // error is an error message from the API. 
        error: undefined
    },
    mutations: {
        error(state, payload) {
            state.error = payload
            Toast.open({
                message: `Boardstate error: ${state.error}`,
                duration: 3000,
                position: "is-bottom",
                type: "is-danger",
            })
        },
        // * updateBoardStates takes an array of boardstates and updates each 
        // boardstate in our local state.
        // * payload is iterated over and checked for it's user's ID
        // and then assigned to the boardstates object keyed by that ID
        updateBoardStates(state, payload) {
            // update each boardstate by player ID
            payload.forEach((bs) => {
                if (bs.User.ID == "" || bs.User.ID == undefined) {
                    state.error = "boardstate did not have ID"
                    return
                }
                state.boardstates[bs.User.ID] = bs
            })
        },
        updateSelf(state, payload) {
            state.self = payload
        },
    },
    actions: {
        // move is a low level )function designed to be a base unit of card 
        // manipulation in our application. 
        // * complex actions should (*eventually) be composed of different moves
        // * move should be atomic.
        // * all card actions should be able to be expressed via a `move` 
        // invocation given the proper arguments.
        // * move only moves cards around a users Boardstate.
        // * if a card needs to move between _players_, `give` should be used.
        move({ state, dispatch }, { userID, src, dest, target })  {
            // snapshot what our current boardstate is so that we're not 
            // mutating state outside of commits
            const self = Object.assign({}, state.boardstates[userID])
            // map all card movements onto that snapshot
            if (!self.User.ID) {
                // require a User.ID to be set or we error
                return dispatch('error', 'failed to find boardstate')
            }
            const from = self[src]
            if (!from) {
                return dispatch('error', 'failed to find target source')
            }
            const to = self[dest]
            if (!to) {
                return dispatch('error', 'failed to find target destination')
            }
            // lookup by the index card.ID
            var selected;
            var foundIdx = from.findIndex(x => x.ID === target)
            if (foundIdx === 0) {
                return dispatch('error', 'failed to find target ', target)
            }
            var selected = from[foundIdx]
            from.splice(foundIdx, 1)
            to.push(selected)

            // assign source and destination
            self[src] = [...from]
            self[dest] = [...to]

            // finally, dispatch our self update as a single action.
            dispatch('mutateBoardState', self)
        },
        // draw will draw a card into hand from the top of the board state's  
        // library. 
        // * if none exists, it errors and declares your loss.
        // * we always treat the "top" of the deck as the card at index 0
        // * thus the bottom of the deck is the nth element of the array
        draw({ commit, dispatch }, boardstate) {
            const bs = Object.assign({}, boardstate)
            if (bs.Library.length < 1) {
                // handle player losing issue
                commit('error', 'you cannot draw from an empty library. you lose the game.')
                // TODO: Make it so that losing triggers a server event.
                // Send the player to the score screen unless they override the 
                // loss.
                return
            }
            const card = bs.Library.shift()
            bs.Hand.push(card)
            dispatch('mutateBoardState', bs)
        },
        // tapAll will tap all cards in the boardstate passed to it.
        // it early returns if there are no cards on the field.
        tapAll({ dispatch }, boardstate) {
            const bs = Object.assign({}, boardstate)
            if (bs.Field.length < 1) {
                // nothing to tap, so just return
                return
            }
            // set tapped to true for each card on the field for the Boardstate
            bs.Field.forEach((card, i) => {
                card.Tapped = true
            })
            // and dispatch the mutation
            dispatch('mutateBoardState', bs)
        },
        // untapAll will untap all cards in the boardstate's Battlefield.
        // it early returns if no cards are present on the player's battlefield
        untapAll({ dispatch }, boardstate) {
            const bs = Object.assign({}, boardstate)
            if (bs.Field.length < 1) {
                // nothing to tap, so just return
                return
            }
            // set tapped to true for each card on the field for the Boardstate
            bs.Field.forEach((card, i) => {
                card.Tapped = false
            })
            // and dispatch the mutation
            dispatch('mutateBoardState', bs)
        },
        // mutate boardstate will mutate a boardstate and then commit the result
        // * this is used as a more low level action and is usually called in 
        // other actions after a Boardstate has been copied and safely mutated.
        // * this can be used directly by the board but actions are cleaner and
        // are the recommended way of altering the board state.
        mutateBoardState({ commit }, payload) {
            api.mutate({
                mutation: updateBoardStateQuery,
                variables: {
                    boardstate: payload,
                },
            })
                .then((resp) => {
                    commit('updateBoardStates', [resp.data.updateBoardState])
                })
                .catch((err) => {
                    console.error('mutateBoardState: error updating boardstate: ', err)
                    commit('error', 'something went wrong.')
                })
        },
        // subAll fetches all boardstates from the server and then subscribes 
        // to all of them.
        subAllBoardstates({ commit, dispatch, rootState }, {gameID, obsID}) {
            return new Promise((resolve, reject) => {
                api.query({
                    query: boardstates,
                    variables: {
                        gameID: gameID
                    }
                })
                .then((resp) => {
                    resp.data.boardstates.forEach((boardstate) => {
                        // dispatch boardstate subscription here
                        dispatch('subToBoardstate', {
                            obsID: obsID,
                            userID: boardstate.User.ID,
                        })
                        if (boardstate.User.ID === rootState.Users.User.ID) {
                            commit('updateSelf', boardstate)
                        } else {
                            commit('updateBoardStates', resp.data.boardstates)
                        }
                    })
                    // Note: we don't need to put this resp into an array 
                    // because its already a list
                    return resolve(resp.data)
                })
                .catch((err) => {
                    console.error('GET BOARDSTATES FAILED: ', err)
                    return reject(err)
                })
            })
        },
        // used for subscribing to single boardstate updates. the observer ID 
        // is the current user's ID and the userID is the ID of the user whose 
        // boardstate is being subscribed.
        subToBoardstate({ rootState, commit }, payload) {
            const sub = api.subscribe({
                query: boardstateSubscription,
                variables: {
                    obsID: payload.obsID,
                    userID: payload.userID,
                },
            })
            sub.subscribe({
                next(data) {
                    // detect self vs opponents here and assign accordingly 
                    if (data.data.boardstateUpdated.User.ID == rootState.Users.User.ID) {
                        commit('updateSelf', data.data.boardstateUpdated)
                    }
                    console.log('boardstate subscription received: ', data.data.boardstateUpdated)
                    commit('updateBoardStates', [data.data.boardstateUpdated])
                },
                error(err) {
                    commit('error', err)
                    console.error('subscribeToBoardstate: boardstate subscription error: ', err)
                }
            })
        },
    },
}

export const Games = {
    state: {
        // game should be identical in structure to the object we get 
        // back from the server. Thus the capitalization.
        game: {
            ID: "",
            Turn: {
                Player: "",
                Phase: "",
                Number: 0
            },
            PlayerIDs: []
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
            // merge updated game over current game
            const g = Object.assign(state.game, game)
            state.game = g
        },
        updateTurn(state, turn) {
            state.game.Turn = turn
        },
        gameFailure(state, error) {
            state.error = error
        },
    },
    actions: {
        getGame({ commit }, ID) {
            return new Promise((resolve, reject) => {
                api.query({
                    query: gameQuery,
                    variables: {
                        gameID: ID,
                    }
                }).then((resp) => {
                    commit('updateGame', resp.data.games[0])
                    return resolve(resp.data.games[0])
                }).catch((err) => {
                    console.error('vuex failed to get game: ', err)
                    commit('gameFailure', err)
                    return reject(err)
                })
            })
        },
        // subscribesToGame takes a gameID and a userID and creates a new 
        // subscription to the Game.
        subscribeToGame({ state, commit, dispatch }, { gameID, userID }) {
            api.query({
                query: gameQuery,
                variables: {
                    gameID: gameID,
                    userID: userID,
                }
            })
            .then(data => {
                if (data.data.games.length === 0) {
                    commit('error', 'no game received from subscription')
                    return
                }
                // update the game that we receive and then subscribe to 
                // updates 
                commit('updateGame', data.data.games[0])
                const sub = api.subscribe({
                    query: gameUpdateQuery,
                    variables: {
                        gameID: gameID,
                        userID: userID,
                    }
                })
                sub.subscribe({
                    next(data) {
                        const g = data.data.gameUpdated
                        // check if any players joined and sub to their board
                        // states
                        if (g.PlayerIDs.length > state.game.PlayerIDs.length) {
                            for (const player in g.PlayerIDs) {
                                dispatch('subToBoardstate', {
                                    userID: player,
                                    obsID: player,
                                })
                            }
                        }
                        // commit the received game updates
                        commit('updateGame', data.data.gameUpdated)
                    },
                    error(err) {
                        console.error('vuex error: subscribeToGame: game subscription error: ', err)
                        commit('error', err)
                    }
                })
            })
        },
        joinGame({ commit }, payload) {
            return new Promise((resolve, reject) => {
                api.mutate({
                    mutation: gql`mutation ($InputJoinGame: InputJoinGame) {
                    joinGame(input: $InputJoinGame) {
                        ID
                        PlayerIDs {
                        Username
                        ID
                        }
                        Turn {
                            Phase
                            Player
                            Number
                        }
                    }
                    }`,
                    variables: {
                        InputJoinGame: payload.inputGame,
                    }
                })
                .then((res) => {
                    commit('updateGame', res.data.joinGame)
                    router.push({ path: `/games/${res.data.joinGame.ID}` })
                    return resolve(res)
                })
                .catch((err) => {
                    commit('error', 'error joining game')
                    console.error('error joining game: ', err)
                    return reject(err)
                })
            })
        },
        createGame({ commit }, payload) {
            commit('loading', true)
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
                        PlayerIDs {
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
                    commit('updateGame', res.data.createGame)
                    router.push({ path: `/games/${res.data.createGame.ID}` })
                    return resolve(res)
                })
                .catch((err) => {
                    commit('error', 'error creating game')
                    console.error('error createGame: ', err)
                    return reject(err)
                })
            })
        },
        updateGame({ commit }, payload) {
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
        }
    },
}

export const Users = {
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
            state.User.Username = payload.Username
            Cookies.set("username", payload.Username)
            state.User.ID = payload.ID
            Cookies.set("userID", payload.ID)
            state.User.Token = payload.Token
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
        searchByName({commit}, name) {
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
    Boardstates,
    Cards,
    Games,
    Users,
  }
})
