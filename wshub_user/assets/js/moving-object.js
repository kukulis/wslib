export class MovingObject {
    /** @type {string} */
    id = '';
    /** @type {number} */
    latitude = 0;
    /** @type {number} */
    longitude = 0;

    /**
     * @param {{id: string, latitude: number, longitude: number}} dto an anonymous object, as parsed from JSON
     * @returns {MovingObject}
     */
    fromDto(dto) {
        this.id = dto.id;
        this.latitude = dto.latitude;
        this.longitude = dto.longitude;
        return this;
    }
}
