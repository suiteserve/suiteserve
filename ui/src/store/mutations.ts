import {MutationTree} from 'vuex';
import * as api from '@/api';
import state from '@/store/state';

interface VersionedEntities<T extends api.VersionedEntity> {
  pendingUpdates: ReadonlyArray<api.UpdateChange<T>>,
  entities: ReadonlyArray<T>,
}

function versionedInsert<T extends api.VersionedEntity>(
    state: VersionedEntities<T>,
    entity: T,
): void {
  // use the newest of the existing and given entity, if applicable
  const existing = state.entities
      .find(e => e.id === entity.id);
  if (existing && existing.version > entity.version) {
    entity = existing;
  }

  // apply any pending update
  const pending = state.pendingUpdates
      .find(e => e.id === entity.id);
  entity = api.applyUpdateChange(entity, pending);

  state.pendingUpdates = state.pendingUpdates
      .filter(e => e.id !== entity.id);
  state.entities = state.entities
      .filter(e => e.id !== entity.id)
      .concat(entity);
}

function update<T extends api.VersionedEntity>(
    state: VersionedEntities<T>,
    update: api.UpdateChange<T>,
): void {
  const toUpdate = state.entities
      .find(e => e.id === update.id);
  if (toUpdate) {
    // apply the update
    state.entities = state.entities
        .filter(e => e.id !== update.id)
        .concat(api.applyUpdateChange(toUpdate, update));
  } else {
    // queue the update
    const pending = state.pendingUpdates
        .find(e => e.id === update.id);
    state.pendingUpdates = state.pendingUpdates
        .filter(e => e.id !== update.id)
        .concat(api.mergeUpdateChanges(pending, update));
  }
}

const suiteAggsId = 'suite_aggs';

export default <MutationTree<typeof state>>{
  insertSuitePage(state, page: api.SuitePage): void {
    versionedInsert(state.suites.aggs, {
      ...page.aggs,
      id: suiteAggsId,
    });
    page.suites?.forEach(s => versionedInsert(state.suites, s))
    state.suites.nextId = page.next_id;
    state.suites.init = true;
  },
  insertSuite(state, s: api.Suite): void {
    versionedInsert(state.suites, s);
    state.suites.init = true;
  },
  updateSuite(state, u: api.UpdateChange<api.Suite>): void {
    update(state.suites, u);
    state.suites.init = true;
  },
  insertSuiteAggs(state, a: api.SuiteAggs): void {
    versionedInsert(state.suites.aggs, {
      ...a,
      id: suiteAggsId,
    });
  },
  updateSuiteAggs(state, a: api.UpdateChange<api.SuiteAggs>): void {
    update(state.suites.aggs, {
      ...a,
      id: suiteAggsId,
    });
  },
};
