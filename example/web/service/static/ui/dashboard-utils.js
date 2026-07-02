(function (global) {
    function safeArray(value) {
        return Array.isArray(value) ? value : [];
    }

    function parseDate(value) {
        if (!value || value === '0001-01-01 00:00:00') {
            return 0;
        }
        const normalized = String(value).replace(' ', 'T');
        const time = new Date(normalized).getTime();
        return Number.isNaN(time) ? 0 : time;
    }

    function normalizeOnlineValue(value) {
        if (typeof value === 'boolean') {
            return value;
        }
        if (typeof value === 'number') {
            return value > 0;
        }
        if (typeof value === 'string') {
            const normalized = value.trim().toLowerCase();
            if (!normalized) {
                return false;
            }
            return ['1', 'true', 'online', 'on', 'yes', 'y', '在线'].indexOf(normalized) >= 0;
        }
        return false;
    }

    function normalizeVehicle(item) {
        const source = item && typeof item === 'object' ? item : {};
        let onlineValue = source.isOnline;
        if (onlineValue === undefined || onlineValue === null || onlineValue === '') {
            onlineValue = source.online;
        }
        if (onlineValue === undefined || onlineValue === null || onlineValue === '') {
            onlineValue = source.onlineStatus;
        }
        let isOnline = normalizeOnlineValue(onlineValue);
        if (onlineValue === undefined || onlineValue === null || onlineValue === '') {
            isOnline = parseDate(source.joinTime) > parseDate(source.leaveTime);
        }
        return Object.assign({}, source, {
            isOnline: isOnline,
            images: safeArray(source.images),
            minioUrls: safeArray(source.minioUrls),
            messages: safeArray(source.messages)
        });
    }

    function normalizeVehicles(list) {
        return Array.isArray(list) ? list.map(normalizeVehicle) : [];
    }

    function getOnlineRatePercent(online, total) {
        if (!total) {
            return 0;
        }
        const rate = Number(((online / total) * 100).toFixed(1));
        if (rate < 0) {
            return 0;
        }
        if (rate > 100) {
            return 100;
        }
        return rate;
    }

    function getOnlineRateText(online, total) {
        const rate = getOnlineRatePercent(online, total);
        if (Number.isInteger(rate)) {
            return rate + '%';
        }
        return rate.toFixed(1).replace(/\.0$/, '') + '%';
    }

    global.DashboardUtils = {
        parseDate: parseDate,
        normalizeOnlineValue: normalizeOnlineValue,
        normalizeVehicle: normalizeVehicle,
        normalizeVehicles: normalizeVehicles,
        getOnlineRatePercent: getOnlineRatePercent,
        getOnlineRateText: getOnlineRateText
    };
})(typeof globalThis !== 'undefined' ? globalThis : this);
