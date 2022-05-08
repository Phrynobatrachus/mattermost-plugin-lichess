import React from 'react';
import {ChannelHeaderButtonIcon} from './icons';

export default class LichessPlugin {
    initialize(registry, store) {
        registry.registerChannelHeaderButtonAction(
            <ChannelHeaderButtonIcon/>,
            () => {},
            'Login to Lichess',
            'Login to Lichess',
        );
    }
}