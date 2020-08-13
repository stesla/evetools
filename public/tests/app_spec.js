describe('evetools', function() {
  let viewContainer;
  beforeEach(function() {
    viewContainer = document.querySelector('.view-container');
  });

  it('can show the login view', function() {
    evetools.showView('');
    expect(viewContainer.querySelectorAll('.view-login').length).toEqual(1);
  });

  it('can show the home view', function() {
    evetools.showView('', {});
    expect(viewContainer.querySelectorAll('.view-home').length).toEqual(1);
  });
});
