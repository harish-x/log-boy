const { Client } = require("@elastic/elasticsearch");
require("dotenv").config();

const client = new Client({
  node: process.env.ElasticSearch,
});
client
  .ping({ requestTimeout: 30000 })
  .then(() => console.log("Elasticsearch is running"))
  .catch((e) => console.log(e));

module.exports = client;
