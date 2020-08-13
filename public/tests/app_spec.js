describe('evetools', function() {
  let viewContainer;
  beforeEach(function() {
    viewContainer = document.querySelector('.view-container');
  });

  it('shows the branding', function() {
    expect(document.querySelectorAll('.branding').length).toEqual(1);
  });

  it('can unhide the main section', function() {
    let main = document.querySelector('section.main');
    expect(main.classList).toContain("hidden");
    evetools.showView('');
    expect(main.classList).not.toContain("hidden");
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
