import { VizPanel } from '../components';
import { Scene } from '../components/Scene';
import { SceneCanvasText } from '../components/SceneCanvasText';
import { SceneSubMenu } from '../components/SceneSubMenu';
import { SceneTimePicker } from '../components/SceneTimePicker';
import { SceneFlexLayout } from '../components/layout/SceneFlexLayout';
import { SceneTimeRange } from '../core/SceneTimeRange';
import { VariableValueSelectors } from '../variables/components/VariableValueSelectors';
import { SceneVariableSet } from '../variables/sets/SceneVariableSet';
import { TestVariable } from '../variables/variants/TestVariable';

import { getQueryRunnerWithRandomWalkQuery } from './queries';

export function getVariablesDemo(): Scene {
  const scene = new Scene({
    title: 'Variables',
    $variables: new SceneVariableSet({
      variables: [
        new TestVariable({
          name: 'server',
          query: 'A.*',
          value: 'server',
          text: '',
          delayMs: 1000,
          options: [],
        }),
        new TestVariable({
          name: 'pod',
          query: 'A.$server.*',
          value: 'pod',
          delayMs: 1000,
          isMulti: true,
          text: '',
          options: [],
        }),
        new TestVariable({
          name: 'handler',
          query: 'A.$server.$pod.*',
          value: 'handler',
          delayMs: 1000,
          //isMulti: true,
          text: '',
          options: [],
        }),
      ],
    }),
    body: new SceneFlexLayout({
      direction: 'row',
      children: [
        new SceneFlexLayout({
          children: [
            new VizPanel({
              pluginId: 'timeseries',
              title: 'handler: $handler',
              $data: getQueryRunnerWithRandomWalkQuery({
                alias: 'handler: $handler',
              }),
            }),
            new SceneCanvasText({
              layout: { width: '40%' },
              text: 'server: ${server} pod:${pod}',
              fontSize: 20,
              align: 'center',
            }),
          ],
        }),
      ],
    }),
    $timeRange: new SceneTimeRange(),
    actions: [new SceneTimePicker({})],
    subMenu: new SceneSubMenu({
      children: [new VariableValueSelectors({})],
    }),
  });

  return scene;
}
