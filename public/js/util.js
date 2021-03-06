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

  const leadingZero = n => n > 9 ? n : '0' + n;

  window.formatDate = function(str) {
    let date = new Date(str);
    return date.getFullYear() + '-' +
        leadingZero(date.getMonth() + 1) + '-' +
        leadingZero(date.getDate()) + ' ' +
        leadingZero(date.getHours()) + ':' +
        leadingZero(date.getMinutes()) + ':' +
        leadingZero(date.getSeconds());
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

  window.byTypeName = function(a, b) {
    return a.type.name < b.type.name ? -1 : 1;
  }

  window.chartPoint = function(d) {
    return {
      date: Date.parse(d.date),
      average: +d.average,
    }
  }
})(window, document);
