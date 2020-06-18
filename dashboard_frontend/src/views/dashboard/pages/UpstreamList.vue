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
                >
              <template v-slot:item.rtt="{ item }">
                <span :class="getServiceHealthTextClasses(item.url)">{{ getServiceHealthText(item.url) }}</span>
              </template>
              <template v-slot:no-data>
                <v-alert
                  :value="true"
                  color="error"
                  icon="warning"
                  >
                  No Services
                </v-alert>
              </template>
            </v-data-table>
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
                >
              <template v-slot:item.rtt="{ item }">
                <span :class="getServiceHealthTextClasses(item.url)">{{ getServiceHealthText(item.url) }}</span>
              </template>
              <template v-slot:no-data>
                <v-alert
                  :value="true"
                  color="error"
                  icon="warning"
                  >
                  No Services
                </v-alert>
              </template>
            </v-data-table>
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
        },
        {
          sortable: false,
          text: "Ping",
          value: "rtt"
        },
      ],
      healths: {},
    };
  },
  computed: {
    ...mapGetters(["upstreamList", "serviceList"])
  },
  mounted () {
      this.$store.dispatch('loadStatistic')
        .then(res => {
          console.log('statistic', res)
          const services = this.serviceList
          console.log('services', services)
          for(const service of services) {
            this.$store.dispatch('queryServiceHealthByUrl', service.url)
              .then(health => {
                console.log(`service ${service.url} health status`, health)
                this.$set(this.healths, service.url, health)
              })
          }
        })
  },
  methods: {
    getServiceHealthTextClasses(serviceUrl) {
      const text = this.getServiceHealthText(serviceUrl)
      if(text==='not connected') {
        return '-not-connected -ping-text'
      } else {
        return '-ping-text'
      }
    },
    getServiceHealthText(serviceUrl) {
      const health = this.healths[serviceUrl]
      if(!health) {
        return 'not connected'
      } else {
        return `${health.rtt} ms`
      }
    }
  }
};
</script>

<style scoped>
.-ping-text {
  color: #aaaaaa;
}
.-ping-text.-not-connected {
  color: red;
}
</style>