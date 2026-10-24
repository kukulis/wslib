/**
 * @param name
 * @returns {HTMLElement}
 */
export function NewE(name) {
    return document.createElement(name)
}

/**
 * @param {String} text
 * @returns {Text}
 */
export function NewT(text) {
    return document.createTextNode(text)
}

export function NewECT(name, className, text) {
    const element = NewE(name);

    if ( className) {
        element.classList.add(className);
    }
    element.appendChild(NewT(text));

    return element;
}

export function NewEC(name, className) {
    const element = NewE(name);
    element.classList.add(className);

    return element;
}

/**
 *
 * @param {string} id
 * @returns {HTMLElement}
 */
export function GetE(id) {
    return document.getElementById(id)
}

/**
 * @param {string} name
 * @returns {string|null}
 */
export function GetCookie(name) {
    for (const pair of document.cookie.split('; ')) {
        const [key, value] = pair.split('=');
        if (key === name) {
            return decodeURIComponent(value);
        }
    }

    return null;
}

/**
 * @param {HTMLElement} element
 * @returns {HTMLElement}
 */
export function ClearE(element) {
    while (element.firstChild) {
        element.removeChild(element.lastChild);
    }

    return element
}

export function AppChild(parent, child) {
    parent.appendChild(child);

    return parent;
}
export function AppChildren(parent, children) {
    for (const child of children) {
        parent.appendChild(child);
    }

    return parent;
}

export function pad(number, digits) {
    let s = "000000000" + number;
    return s.substring(s.length-digits);
}

/**
 * @param {number} priceInCents
 * @returns {string}
 */
export function formatPrice(priceInCents) {
    return (priceInCents / 100).toFixed(2);
}

/**
 * @param {string} tag
 * @param {string} className
 * @param {string} text
 * @returns {HTMLElement}
 */
export function Cell(tag, className, text) {
    const cell = NewE(tag);
    cell.classList.add(className);
    cell.appendChild(NewT(text));

    return cell;
}

/**
 * @param {string} className
 * @param {string} text
 * @param {string} href
 * @returns {HTMLElement}
 */
export function LinkCell(className, text, href) {
    const cell = NewE('td');
    cell.classList.add(className);

    const link = AppChild(NewE('a'), NewT(text));
    link.href = href;
    cell.appendChild(link);

    return cell;
}

/**
 * @param {string} className
 * @param {string} symbol
 * @param {() => void} onClick
 * @returns {HTMLElement}
 */
export function ActionCell(className, symbol, onClick) {
    const cell = NewE('td');
    cell.classList.add(className);

    const link = AppChild(NewE('a'), NewT(symbol));
    link.href = '#';
    link.addEventListener('click', (event) => {
        event.preventDefault();
        onClick();
    });
    cell.appendChild(link);

    return cell;
}
