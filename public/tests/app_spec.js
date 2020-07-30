describe('evetools', function() {
  var viewContainer;
  beforeEach(function() {
    viewContainer = document.querySelector('.view-container');
  });

  it('shows the branding', function() {
    expect(document.querySelectorAll('.branding').length).toEqual(1);
  });

  it('can show the landing view', function() {
    evetools.showView('');
    expect(viewContainer.querySelectorAll('.landing-view').length).toEqual(1);
  });

  it('can show the profile view', function() {
    evetools.showView('', {});
    expect(viewContainer.querySelectorAll('.profile-view').length).toEqual(1);
  });
});
