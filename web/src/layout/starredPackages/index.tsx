import isNull from 'lodash/isNull';
import isUndefined from 'lodash/isUndefined';
import React, { useContext, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';

import { API } from '../../api';
import { AppCtx, signOut } from '../../context/AppCtx';
import { Package } from '../../types';
import Loading from '../common/Loading';
import NoData from '../common/NoData';
import PackageCard from '../common/PackageCard';
import styles from './StarredPackagesView.module.css';

const StarredPackagesView = () => {
  const history = useHistory();
  const { dispatch } = useContext(AppCtx);
  const [isLoading, setIsLoading] = useState(false);
  const [packages, setPackages] = useState<Package[] | undefined>(undefined);
  const [apiError, setApiError] = useState<string | JSX.Element | null>(null);

  const onAuthError = (): void => {
    dispatch(signOut());
    history.push('/login?redirect=/user/packages/starred');
  };

  useEffect(() => {
    async function fetchStarredPackages() {
      try {
        setIsLoading(true);
        setPackages(await API.getStarredByUser());
        setApiError(null);
        setIsLoading(false);
      } catch (err) {
        setIsLoading(false);
        if (err.statusText !== 'ErrLoginRedirect') {
          setApiError('An error occurred getting your starred packages, please try again later');
          setPackages([]);
        } else {
          onAuthError();
        }
      }
    }
    fetchStarredPackages();
  }, []); /* eslint-disable-line react-hooks/exhaustive-deps */

  return (
    <div className="d-flex flex-column flex-grow-1 position-relative">
      {(isUndefined(packages) || isLoading) && <Loading />}

      <main role="main" className="container py-5">
        <div className="flex-grow-1 position-relative">
          <div className="h4 pb-0">
            <div className="d-flex align-items-center justify-content-center">
              <div>Your starred packages</div>
            </div>
          </div>

          <div className={`mx-auto mt-4 ${styles.wrapper}`}>
            {!isUndefined(packages) && (
              <>
                {packages.length === 0 ? (
                  <NoData issuesLinkVisible={!isNull(apiError)}>
                    {isNull(apiError) ? <>You have not starred any package yet</> : <>{apiError}</>}
                  </NoData>
                ) : (
                  <>
                    {packages.map((item: Package) => (
                      <PackageCard key={item.packageId} package={item} fromStarredPage={true} />
                    ))}
                  </>
                )}
              </>
            )}
          </div>
        </div>
      </main>
    </div>
  );
};

export default StarredPackagesView;
