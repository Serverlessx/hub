import { SearchFiltersURL } from '../types';
import buildSearchParams from './buildSearchParams';

interface Test {
  query: string;
  result: SearchFiltersURL;
}

const tests: Test[] = [
  {
    query: '?page=1&deprecated=true&org=org',
    result: {
      pageNumber: 1,
      filters: {
        org: ['org'],
      },
      deprecated: true,
    },
  },
  {
    query: '?page=1&text=test',
    result: {
      pageNumber: 1,
      text: 'test',
      filters: {},
      deprecated: false,
    },
  },
  {
    query: '?text=test',
    result: {
      pageNumber: 1,
      text: 'test',
      filters: {},
      deprecated: false,
    },
  },
  {
    query: '?page=1',
    result: {
      pageNumber: 1,
      filters: {},
      deprecated: false,
    },
  },
  {
    query: '?page=3&deprecated=true&org=org1&org=org2&repo=stable&repo=incubator',
    result: {
      pageNumber: 3,
      filters: {
        org: ['org1', 'org2'],
        repo: ['stable', 'incubator'],
      },
      deprecated: true,
    },
  },
  {
    query: '?page=10&deprecated=true&user=user1&user=user2',
    result: {
      pageNumber: 10,
      filters: {
        user: ['user1', 'user2'],
      },
      deprecated: true,
    },
  },
];

describe('buildSearchParams', () => {
  for (let i = 0; i < tests.length; i++) {
    it('renders proper filters', () => {
      const actual = buildSearchParams(tests[i].query);
      expect(actual).toEqual(tests[i].result);
    });
  }
});
