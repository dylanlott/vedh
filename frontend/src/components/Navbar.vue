<template>
    <b-navbar>
        <template slot="brand">
            <b-navbar-item tag="router-link" :to="{ path: '/' }">
                <!-- TODO: Replace with visual logo -->
                <!-- <img src="../logo.png" alt="EDH Go"> -->
               <h1 class="title"><b>vEDH.xyz</b></h1>
            </b-navbar-item>
        </template>
        <template slot="start">
            <b-navbar-item href="#">
                Home
            </b-navbar-item>
            <b-navbar-item href="/games">
                Games
            </b-navbar-item>
            <b-navbar-item href="/join">
                Join
            </b-navbar-item>
        </template>

        <template slot="end">
            <b-navbar-item tag="div">
                <div class="buttons" v-if="!user.Token">
                    <a href="/signup" class="button is-primary">
                        <strong>Sign up</strong>
                    </a>
                    <a tag="router-link"
                    href="/login"
                    class="button is-light">
                        Log in
                    </a>
                </div>
                <div v-else>
                    <!-- ### TODO: Create the Profile page ###
                    <a tag="router-link"
                    href="/profile"
                    class="button is-dark">
                       Profile 
                    </a> -->
                    <a tag="router-link"
                    href="#"
                    @click="handleLogout()"
                    class="button is-light">
                       Log Out 
                    </a>
                </div>
            </b-navbar-item>
        </template>
    </b-navbar>
</template>
<script>
import { mapState } from 'vuex';

export default {
  name: 'navbar',
  computed: mapState({
      user: state => state.Users.User
  }),
  methods: {
      handleLogout() {
          this.$store.dispatch('Users/logout')
      }
  }
}
</script>
