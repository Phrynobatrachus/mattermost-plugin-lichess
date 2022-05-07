import React from 'react';

export default class LichessPlugin {
    initialize(registry, store) {
        registry.registerCustomRoute(
            '/login',
            () => 'login route'
        )
    }
}