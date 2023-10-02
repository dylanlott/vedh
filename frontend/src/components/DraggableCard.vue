<template>
  <div ref="myDraggable" class="draggable">
    <!--  Wrap Card here -->
    <Card v-bind="card" :ScreenX="screenX" v-bind:ScreenY="screenY" />
  </div>
</template>

<script>
import interact from 'interactjs';
import Card from '@/components/Card';
export default {
  name: 'DraggableCard',
  data() {
    return {
      screenX: 0,
      screenY: 0,
      lastSource: '',
      lastDestination: '',
    };
  },
  props: {
    card: Object,
    user: Object,
  },
  mounted: function () {
    let myDraggable = this.$refs.myDraggable;
    this.initInteract(myDraggable);
  },
  methods: {
    initInteract: function (selector) {
      interact(selector).draggable({
        // enable inertial throwing
        inertia: true,
        // enable autoScroll
        autoScroll: true,
        // call this function on every dragmove event
        onmove: this.dragMoveListener,
        // call this function on every dragend event
        onend: this.onDragEnd,
      });
    },
    dragMoveListener: function (event) {
      var target = event.target,
        // keep the dragged position in the data-x/data-y attributes
        x = (parseFloat(target.getAttribute('data-x')) || this.screenX) + event.dx,
        y = (parseFloat(target.getAttribute('data-y')) || this.screenY) + event.dy;

      // translate the element
      target.style.webkitTransform = target.style.transform = 'translate(' + x + 'px, ' + y + 'px)';

      // update the position attributes
      target.setAttribute('data-x', x);
      target.setAttribute('data-y', y);

      if (!!event.dragLeave) {
        this.lastSource = event.dragLeave.id
      }
      if (!!event.dragEnter) {
        this.lastDestination = event.dragEnter.id
        this.card.CurrentZone = this.lastDestination
      }
    },
    onDragEnd: function (event) {
      var target = event.target;
      this.screenX = target.getBoundingClientRect().left;
      this.screenY = target.getBoundingClientRect().top;
      this.card.ScreenX = this.screenX
      this.card.ScreenY = this.screenY

      let game = this.$store.state.Games.game
      let userID = this.$store.state.Users.User.ID
      let self = game.Players.find(x => x.ID === userID)
      if (self === undefined) {
        console.log("cannot find self: investigate this problem", game, userID)
        return
      }
      if (!!this.lastSource && !!this.lastDestination) {
        let sourceArray = self.Boardstate[this.lastSource]
        let targetArray = self.Boardstate[this.lastDestination]

        let moved = this.move({
          sourceArray: sourceArray,
          targetArray: targetArray,
          card: this.card,
        })

        if (moved) {
          console.log("moved - updated state: ", self)
          // let found = this.$store.state.Games.game.Players.find(x => x.ID === userID) 
          // found = self
          this.$store.dispatch('Games/sync', this.$store.state.Games.game)
        } else {
          console.log("not moved - current state: ", self)
        }
      }
    },
    capitalize: function (word) {
      return word[0].toUpperCase() + word.slice(1).toLowerCase();
    },
    /**
     * Moves an object with a specific ID from one array to another.
     * 
     * @param {Array} sourceArray - The array to remove the object from.
     * @param {Array} targetArray - The array to add the object to.
     * @param {string|number} id - The ID of the object to move.
     * @returns {boolean} - Returns true if the object was moved successfully, false otherwise.
     */
    move({ sourceArray, targetArray, card }) {
      console.log('moving ', card.ID, ' from ', sourceArray, ' to ', targetArray)
      const index = sourceArray.findIndex(obj => obj.ID === card.ID);
      if (index !== -1) {
        const [movedObject] = sourceArray.splice(index, 1);
        targetArray.push(movedObject);
        return true;
      }
      return false;
    },
    checkOverlap(elem1, elem2) {
      const rect1 = elem1.getBoundingClientRect();
      const rect2 = elem2.getBoundingClientRect();
      return !(rect2.left > rect1.right ||
        rect2.right < rect1.left ||
        rect2.top > rect1.bottom ||
        rect2.bottom < rect1.top);
    },
  },
  components: {
    Card,
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.draggable {
  padding: 5px;
  position: absolute;
}
</style>
