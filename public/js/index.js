viewData = (function(window, document, undefined) {
  var stations = retrieve('/data/stations.json', 'error fetching stations');
  var corps = retrieve('/data/corporations.json', 'error fetching corporations');
  var orders = retrieve('/api/v1/user/orders', 'error fetching market orders');
  var skills = retrieve('/api/v1/user/skills', 'error fetching skills');
  var standings = retrieve('/api/v1/user/standings', 'error fetching standings');
  var wallet = retrieve('/api/v1/user/walletBalance', 'error fetching wallet balance');

  const brokerRelationsID = 3446;

  function calculateBrokerFee(user, stations, corps, skills, standings) {
    // 5%-(0.3%*BrokerRelationsLevel)-(0.03%*FactionStanding)-(0.02%*CorpStanding)
    var fee = 0.05;
    let brokerRelations = skills.find(s => s.skill_id === brokerRelationsID);
    let station = stations[user.station_id];
    let corp = corps[station.corp_id]
    let corpStanding = standings.find(s =>
      s.from_type === 'npc_corp' && s.from_id == corp.id);
    let factionStanding = standings.find(s =>
      s.from_type === 'faction' && s.from_id == corp.faction_id);
    
    if (brokerRelations)
      fee -= 0.003 * brokerRelations.active_skill_level;
    if (factionStanding)
      fee -= 0.0003 * factionStanding.standing;
    if (corpStanding)
      fee -= 0.0002 * corpStanding.standing;

    return fee;
  }

  return {
    data: undefined,
    favorites: [],
    walletBalance: 0,
    brokerFee: 0,
    buyTotal: 0,
    sellTotal: 0,

    initialize() {
      document.title += " - Dashboard"

      wallet.then(balance => {
        this.walletBalance = balance;
      });

      orders.then(orders => {
        this.buyTotal = orders.buy.reduce((a, o) => a + o.escrow, 0);
        this.sellTotal = orders.sell.reduce((a, x) => a + x.volume_remain * x.price, 0);
      });

      evetools.sdeTypes().then(types => {
        evetools.currentUser.then(user => {
          this.favorites = user.favorites.map(id => {
            let type = types[""+id];
            type.favorite = true;
            return type;
          }).sort(byName);
        });
      });

      stations.then(stations => {
        corps.then(corps => {corps
          standings.then(standings => {
            skills.then(skills => {
              evetools.currentUser.then(user => {
                this.brokerFee = calculateBrokerFee(user, stations, corps, skills, standings);
              });
            });
          });
        });
      });
    },

    toggleFavorite(type) {
      let val = !type.favorite
      setFavorite(type.id, val)
      .then(() => {
        type.favorite = val
      });
    },
  }
})(window, document, undefined);
