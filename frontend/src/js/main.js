export let currentPage = 1;
export let currentIndex = 0;
export const itemsPerPage = 10;
export let totalItems = 0;

export function switchProfileSelectionAndEditPages() {
    $('#profile-selection-page').toggle();
    $('#profile-edit-page').toggle();
}

export async function fetchListings(direction) {
    try {
        const listings = await window.go.main.App.Return10Listings(currentIndex, direction);
        return listings;
    } catch (error) {
        console.error('Error fetching listings:', error);
        return [];
    }
}

export async function getTotalItemsCount() {
    try {
        return await window.go.main.App.GetTotalItems();
    } catch (error) {
        console.error('Error fetching total count:', error);
        return 0;
    }
}

export async function renderListings(listings) {
    const modsList = $('#mods-list');
    modsList.empty();

    await Promise.all(listings.map(async listing => {
        download_url = await window.go.main.App.GetDownloadURL(listing);
        modsList.append(`
            <li>
                <div class="grid">
                    <span>${listing.name} - ${listing.owner}</span>
                    <button onclick="window.go.main.App.Download('${download_url}')" >Download</button>
                    <button onclick="BrowserOpenURL('${listing.package_url}')" >Open website &rarr;</button>
                </div>
            </li>
        `);
    }));
}

export function updatePaginationUI() {
    $('#page-number').text(currentPage);
    $('#page-prev, #page-first').prop('disabled', currentPage === 1);
    $('#page-next, #page-last').prop('disabled', currentPage === Math.ceil(totalItems / itemsPerPage));
}

export async function changePage(newPage) {
    const direction = newPage > currentPage ? 1 : -1;
    currentIndex = (newPage - 1) * itemsPerPage;

    const listings = await fetchListings(direction);
    if (listings.length > 0) {
        currentPage = newPage;
        renderListings(listings);
        updatePaginationUI();
    }
}
