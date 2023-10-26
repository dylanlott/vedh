<template>
    <div>
        <b-collapse class="card" animation="slide" aria-id="contentIdForA11y3">
            <template #trigger="props">
                <div class="card-header" role="button" aria-controls="contentIdForA11y3" :aria-expanded="props.open">
                    <p class="card-header-title">
                        {{ username }} - [ {{ boardstate.Life }} ]
                    </p>
                    <a class="card-header-icon">
                        <b-icon :icon="props.open ? 'menu-down' : 'menu-up'">
                        </b-icon>
                    </a>
                </div>
            </template>

            <div class="card-content">
                <div class="content">
                    <div>
                        BATTLEFIELD
                        <div v-for="card in boardstate.Battlefield">
                            <Card v-bind="card"/>
                        </div>
                    </div>
                    <div>
                        REVEALED
                        <div v-for="card in boardstate.Revealed">
                            <Card v-bind="card"/>
                        </div>
                    </div>
                    <!-- Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus nec iaculis mauris.
            <a>#buefy</a>. -->
                </div>
            </div>
            <footer class="card-footer">
                <!-- <a class="card-footer-item">Save</a>
          <a class="card-footer-item">Edit</a>
          <a class="card-footer-item">Delete</a> -->
            </footer>
        </b-collapse>
    </div>
</template>

<script>
import Card from '@/components/Card';
import { mapState } from 'vuex';
export default {
    name: 'Opponent',
    data() {
        return {
            open: true,
        }
    },
    props: {
        username: String,
        boardstate: Object,
        id: String,
    },
    components: { Card },
}
</script>