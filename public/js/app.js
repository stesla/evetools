evetools = (function(document, window, undefined) {
  var result = {};

  const plainViews = ['login', 'notFound'];

  var verify = retrieve('/api/v1/verify', 'error verifying auth');

  result.globalState = function() {
    return {
      avatarMenuOpen: false,
      marketMenuOpen: false,
      loggedIn: false,
      navOpen: false,
      character: { name: "", id: 0 },

      get currentView() {
        if (!this.loggedIn) {
          return 'login';
        }

        let path = window.location.pathname;

        if (path === '/')
          return 'index';

        if (path.startsWith('/browse'))
          return 'browse';

        if (path.startsWith('/groups/'))
          return 'groupDetails';

        if (path.startsWith('/history'))
          return 'history';

        if (path.startsWith('/orders'))
          return 'orders';

        if (path.startsWith('/search'))
          return 'search';

        if (path.startsWith('/settings'))
          return 'settings';

        if (path.startsWith('/transactions'))
          return 'transactions';

        if (path.startsWith('/types/'))
          return 'typeDetails';

        return 'notFound';
      },

      initialize() {
        verify.then(user => {
          this.character = {
            id: user.character_id,
            name: user.character_name,
          }
          this.loggedIn = true;
          return user;
        })
        .catch(() => {})
        .then(() => {
          if (!plainViews.includes(this.currentView)) {
            const url = '/js/'+this.currentView+'.js';
            return retrieve(url, 'error fetching viewData', { raw: true });
          }
        })
        .then(resp => resp ? resp.blob() : undefined)
        .then(blob => {
          if (blob) {
            const script = document.createElement('script'),
                  src = URL.createObjectURL(blob);
            script.src = src;
            document.body.appendChild(script);
          }
        })
        .then(() => {
          const url = '/views/'+this.currentView+'.html';
          return retrieve(url, 'error fetching view', { raw: true });
        })
        .then(resp => resp.text())
        .then(html => {
          const parser = new DOMParser();
          const elt = parser.parseFromString(html, 'text/html').querySelector('main');
          const slot = document.querySelector('main');
          slot.parentNode.replaceChild(elt, slot);
        });
      },
    }
  }

  return result;
})(document, window);
