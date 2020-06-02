<template>
  <v-container
    id="requestlist"
    fluid
    tag="section"
    >
    <v-row>
      <v-col
        cols="12"
        md="12"
        >
        <base-material-card
            color="warning"
            class="px-5 py-3"
            >
          <template v-slot:heading>
            <div class="display-2 font-weight-light">Request List</div>
          </template>
          <v-card-text>
            <v-data-table
                :headers="requestListHeaders"
                :items="requestList.items"
                :options.sync="tableOptions"
                :server-items-length="pager.total"
                @update:options="optionsChanged"
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
  name: "RequestList",
  data() {
    return {
      requestListHeaders: [
        {
          sortable: false,
          text: "TraceId",
          value: "traceId"
        },
        {
          sortable: false,
          text: "Method",
          value: "rpcMethodName"
        },
        {
          sortable: false,
          text: "Target",
          value: "targetServer"
        },
        {
          sortable: false,
          text: "Annotation",
          value: "annotation"
        },
        {
          sortable: false,
          text: "Error",
          value: "rpcResponseError"
        },
        // {
        //   sortable: false,
        //   text: "Result",
        //   value: "rpcResponseResult"
        // },
        {
          sortable: false,
          text: "Log Time",
          value: "logTime"
        },
        {
          sortable: false,
          text: "Params",
          value: "rpcRequestParams"
        },
      ],
      tableOptions: {
        page: 1,
        itemsPerPage: 10,
      },
      pager: {
          total: 0,
      }
    };
  },
  computed: {
    ...mapGetters(["requestList"])
  },
  mounted () {
      this.$store.dispatch('loadRequestSpanList', {})
        .then(res => {
            this.pager.total = res.total
        })
  },
  methods: {
      optionsChanged () {
          const options = this.tableOptions
          console.log('options', options)
          this.$store.dispatch('loadRequestSpanList', {offset: (options.page-1) * options.itemsPerPage, limit: options.itemsPerPage})
            .then(res => {
                this.pager.total = res.total
            })
      }
  },
};
</script>

<style lang="less" scoped>
</style>