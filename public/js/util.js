(function(window, document, undefined){
  window.retrieve = function(url, errmsg, options) {
    return fetch(url, options)
    .then(resp => {
      if(!resp.ok) {
        throw new Error(errmsg);
      }
      return options && options.raw ? resp : resp.json();
    });
  }

  window.avatarURL = function(id) {
      return id ? 'https://images.evetech.net/characters/' + id + '/portrait?size=128' : undefined;
  }

  window.byName = function(a, b) {
    return a.name < b.name ? -1 : 1;
  }

  window.hrefGroup = function(id) {
    return '/groups/' + id;
  }

  window.hrefType = function(id) {
    return '/types/' + id;
  },

  window.imgURL = function(type) {
    if (!type) return undefined;
    var imgType = 'icon';
    if (type.name.match(/(Blueprint|Formula)$/))
      imgType = 'bp';
    return 'https://images.evetech.net/types/' + type.id + '/' + imgType + '?size=128';
  }

  window.formatDecimal = function(amt, maxDigit, minDigit) {
    return amt.toLocaleString('en-US', { 
      maximumFractionDigits: maxDigit, 
      minimumFractionDigits: minDigit,
    });
  }

  window.formatISK = function(amt) {
    return formatDecimal(amt, 2, 2);
  }

  window.formatNumber = function(amt) {
    return formatDecimal(amt, 2)
  }

  window.formatPercent = function(amt) {
    return formatDecimal(100 * amt, 2, 2) + "%"
  }

  window.handleSearch = function(q) {
    window.location = '/search?q=' + q;
  }

  window.openTypeInEVE = function(typeID) {
    retrieve('/api/v1/types/'+typeID+'/openInGame', 'error making openInGame API call', {
      raw: true,
      method: 'POST',
    });
  }

  window.removeCharacter = function(cid) {
    retrieve('/api/v1/user/characters/' + cid, 'error deleting character', {
      raw: true,
      method: 'DELETE',
    })
    .then(() => {
      location.reload(true);
    });
  }

  window.setFavorite = function(typeID, val) {
    return retrieve('/api/v1/types/'+typeID+'/favorite', 'error setting favorite', {
      method: 'PUT',
      body: JSON.stringify({favorite: val}),
    });
  }

  window.setOrderFields = function(o, types, stations) {
    let type = types[''+o.type_id]
    let station = stations[''+o.location_id]
    o.type = type;
    o.station_name = station.name;
    return o;
  }

  window.byTypeName = function(a, b) {
    return a.type.name < b.type.name ? -1 : 1;
  }

})(window, document);
