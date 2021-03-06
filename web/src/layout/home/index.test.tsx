import { render, screen, waitFor, waitForElementToBeRemoved } from '@testing-library/react';
import React from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import { mocked } from 'ts-jest/utils';

import { API } from '../../api';
import { Stats } from '../../types';
import HomeView from './index';
jest.mock('../../api');
jest.mock('./SearchTip', () => () => <div />);
jest.mock('./PackagesUpdates', () => () => <div />);

const getMockStats = (fixtureId: string): Stats => {
  return require(`./__fixtures__/index/${fixtureId}.json`) as Stats;
};

const defaultProps = {
  isSearching: true,
  onOauthFailed: false,
};

describe('Home index', () => {
  afterEach(() => {
    jest.resetAllMocks();
  });

  it('creates snapshot', async () => {
    const mockStats = getMockStats('1');
    mocked(API).getStats.mockResolvedValue(mockStats);

    const result = render(
      <Router>
        <HomeView {...defaultProps} />
      </Router>
    );
    expect(result.asFragment()).toMatchSnapshot();
    await waitFor(() => {});
  });

  describe('Render', () => {
    it('renders component', async () => {
      const mockStats = getMockStats('2');
      mocked(API).getStats.mockResolvedValue(mockStats);

      render(
        <Router>
          <HomeView {...defaultProps} />
        </Router>
      );
      expect(API.getStats).toHaveBeenCalledTimes(1);
      await waitFor(() => {});
    });

    it('removes loading spinner after getting package', async () => {
      const mockStats = getMockStats('3');
      mocked(API).getStats.mockResolvedValue(mockStats);

      const props = {
        ...defaultProps,
        isSearching: true,
      };
      render(
        <Router>
          <HomeView {...props} />
        </Router>
      );

      const spinner = await waitForElementToBeRemoved(() => screen.getAllByRole('status'));

      expect(spinner).toBeTruthy();
      await waitFor(() => {});
    });

    it('renders dash symbol when results are 0', async () => {
      const mockStats = getMockStats('4');
      mocked(API).getStats.mockResolvedValue(mockStats);

      const props = {
        ...defaultProps,
        isSearching: true,
      };
      render(
        <Router>
          <HomeView {...props} />
        </Router>
      );

      const emptyStats = await waitFor(() => screen.getAllByText('-'));

      expect(emptyStats).toHaveLength(2);
      await waitFor(() => {});
    });

    it('renders dash symbol when getStats call fails', async () => {
      mocked(API).getStats.mockRejectedValue({ status: 500 });

      const props = {
        ...defaultProps,
        isSearching: true,
      };
      render(
        <Router>
          <HomeView {...props} />
        </Router>
      );

      await waitFor(() => expect(API.getStats).toHaveBeenCalledTimes(1));
      expect(screen.getAllByText('-')).toHaveLength(2);
    });

    it('renders project definition', async () => {
      const mockStats = getMockStats('5');
      mocked(API).getStats.mockResolvedValue(mockStats);

      render(
        <Router>
          <HomeView {...defaultProps} />
        </Router>
      );

      const heading = await waitFor(() => screen.getByRole('heading'));

      expect(heading).toBeInTheDocument();
      expect(heading.innerHTML).toBe('Find, install and publish<br>Kubernetes packages');
      await waitFor(() => {});
    });

    it('renders CNCF info', async () => {
      const mockStats = getMockStats('6');
      mocked(API).getStats.mockResolvedValue(mockStats);

      render(
        <Router>
          <HomeView {...defaultProps} />
        </Router>
      );

      const CNCFInfo = await waitFor(() => screen.getByTestId('CNCFInfo'));

      expect(CNCFInfo).toBeInTheDocument();
      expect(CNCFInfo).toHaveTextContent(
        'Artifact Hub aspires to be a Cloud Native Computing Foundation sandbox project.'
      );
      await waitFor(() => {});
    });
  });
});
