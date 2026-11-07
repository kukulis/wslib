import {MovingObject} from "./moving-object.js";

/**
 * The positions of the moving objects that moved. It may hold all the objects or only some of them.
 */
export class MovingObjectsMessage {
    static TYPE = 'moving_objects';

    /** @type {MovingObject[]} */
    objects = [];

    /**
     * @param {MovingObject} object
     * @returns {MovingObjectsMessage}
     */
    addObject(object) {
        this.objects.push(object);
        return this;
    }

    /**
     * @param {{objects: {id: string, latitude: number, longitude: number}[]}} dto the data of the envelope, as parsed from JSON
     * @returns {MovingObjectsMessage}
     */
    fromDto(dto) {
        this.objects = dto.objects.map(objectDto => new MovingObject().fromDto(objectDto));
        return this;
    }
}
