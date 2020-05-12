<template>
  <v-container
    id="upstreamlist"
    fluid
    tag="section"
    >
    <v-row>
      <v-col
        cols="12"
        md="6"
        >
        <base-material-card
            color="warning"
            class="px-5 py-3"
            >
          <template v-slot:heading>
            <div class="display-2 font-weight-light">Upstream List</div>
          </template>
          <v-card-text>
            <v-data-table
                :headers="serviceListHeaders"
                :items="upstreamList"
                />
          </v-card-text>
        </base-material-card>
      </v-col>

      <v-col
        cols="12"
        md="6"
        >
        <base-material-card
            color="warning"
            class="px-5 py-3"
            >
          <template v-slot:heading>
            <div class="display-2 font-weight-light">Service List</div>
          </template>
          <v-card-text>
            <v-data-table
                :headers="serviceListHeaders"
                :items="serviceList"
                />
          </v-card-text>
        </base-material-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
/*eslint-disable */
import { mapGetters } from "vuex";
export default {
  name: "UpstreamList",
  data() {
    return {
      serviceListHeaders: [
        {
          sortable: false,
          text: "Name",
          value: "name"
        },
        {
          sortable: false,
          text: "Url",
          value: "url"
        }
      ]
    };
  },
  computed: {
    ...mapGetters(["upstreamList", "serviceList"])
  },
  mounted () {
      this.$store.dispatch('loadStatistic')
  }
};
</script>

<style lang="less" scoped>
</style>