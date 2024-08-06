import {
    switchProfileSelectionAndEditPages,
    fetchListings,
    getTotalItemsCount,
    renderListings,
    updatePaginationUI,
    changePage,
    itemsPerPage,
    totalItems,
    currentPage,
    currentIndex,
} from './main.js';

$(document).ready(function () {
    $('#create-profile-button').on('click', function () {
        createProfile(prompt('Enter profile name:'));
    });

    $('#back-to-profiles').on('click', function () {
        switchToProfileSelectionPage();
    });

    $('#apply-mods').on('click', function () {
        applySelectedMods();
    });

    $('#debug-switch').on('click', function () {
        switchProfileSelectionAndEditPages();
    });

    $('#page-first').on('click', () => changePage(1));
    $('#page-prev').on('click', () => changePage(currentPage - 1));
    $('#page-next').on('click', () => changePage(currentPage + 1));
    $('#page-last').on('click', async () => {
        const lastPage = Math.ceil(totalItems / itemsPerPage);
        await changePage(lastPage);
    });
});

(async function init() {
    totalItems = await getTotalItemsCount();
    const initialListings = await fetchListings(1);
    renderListings(initialListings);
    updatePaginationUI();
})();