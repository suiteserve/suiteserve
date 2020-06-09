/**
 * @typedef {Object} Error
 * @property {string} error
 */

/**
 * @typedef {Object} SuiteFailureType
 * @property {string} name
 * @property {?string} description
 */

/**
 * @typedef {Object} SuiteEnvVar
 * @property {string} key
 * @property {*} value
 */

/**
 * @typedef {Object} Suite
 * @property {string} id
 * @property {?string} name
 * @property {?SuiteFailureType[]} failure_types
 * @property {?string[]} tags
 * @property {?SuiteEnvVar[]} env_vars
 * @property {?string[]} attachments
 * @property {number} planned_cases
 * @property {string} status
 * @property {boolean} deleted
 * @property {number} started_at
 * @property {?number} finished_at
 * @property {?number} disconnected_at
 * @property {?number} deleted_at
 */

/**
 * @typedef {Object} SuitePage
 * @property {number} running_count
 * @property {number} finished_count
 * @property {string} next_id
 * @property {Suite[]} suites
 */

/**
 * @param {?string} fromId
 * @param {number} [limit=10]
 * @return {Promise<SuitePage>}
 */
export async function fetchSuites({fromId, limit}) {
  const url = new URL('/v1/suites', window.location.href);
  if (fromId != null) {
    url.searchParams.append('from_id', fromId);
  }
  if (limit == null) {
    limit = 10;
  }
  url.searchParams.append('limit', limit.toString());

  const res = await fetch(url.href);
  /** @type {SuitePage|Error} */
  const json = await res.json();

  if (!res.ok) {
    throw `Error fetching suites: ${json.error}`;
  }
  return json;
}

/**
 * @typedef {Object} CaseLink
 * @property {string} type
 * @property {string} name
 * @property {string} url
 */

/**
 * @typedef {Object} CaseArg
 * @property {string} key
 * @property {*} value
 */

/**
 * @typedef {Object} Case
 * @property {string} id
 * @property {string} suite
 * @property {string} name
 * @property {?string} description
 * @property {?string[]} tags
 * @property {number} num
 * @property {?CaseLink[]} links
 * @property {?CaseArg[]} args
 * @property {?string[]} attachments
 * @property {string} status
 * @property {number} created_at
 * @property {?number} started_at
 * @property {?number} finished_at
 */

/**
 * @param {string} suiteId
 * @param {number=} num
 * @return {Promise<Case[]>}
 */
export async function fetchCases(suiteId, num) {
  suiteId = encodeURIComponent(suiteId);
  const url = new URL(`/v1/suites/${suiteId}/cases`, window.location.href);
  if (num != null) {
    url.searchParams.append('num', num.toString());
  }

  const res = await fetch(url.href);
  /** @type {Case[]|Error} */
  const json = await res.json();

  if (!res.ok) {
    throw `Error fetching cases: ${json.error}`;
  }
  return json;
}
