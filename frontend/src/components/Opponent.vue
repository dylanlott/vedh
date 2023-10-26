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
                <div class="opponent-battlefield" v-for="(card, idx) in boardstate.Battlefield">
                    <Card :key="idx" v-bind="card" />
                </div>
            </div>
            <!-- <footer class="card-footer">
            </footer> -->
        </b-collapse>
    </div>
</template>

<script>
import Card from '@/components/Card';
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
<style>
.opponent-battlefield {
    display: inline-block;
}
</style>